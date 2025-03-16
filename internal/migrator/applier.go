package migrator

import (
	"context"
	"fmt"
	"log"
	"math"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/servletcloud/Andmerada/internal/migrator/sqlres"
	"github.com/servletcloud/Andmerada/internal/project"
	"github.com/servletcloud/Andmerada/internal/source"
)

type Report struct {
	SourcesOnDisk  int
	PendingSources int
}

type Applier struct {
	MaxSQLFileSize int64
	DatabaseURL    string
	Project        project.Project
}

func (applier *Applier) ApplyPending(ctx context.Context, report *Report) error {
	sourceIDToName := make(map[source.MigrationID]string)
	dupeIDToName := make(map[source.MigrationID]string)
	idMin, idMax := source.MigrationID(math.MaxUint64), source.MigrationID(0)

	err := source.ScanAll(applier.Project.Dir, func(id source.MigrationID, name string) {
		if _, found := sourceIDToName[id]; found {
			dupeIDToName[id] = name
		} else {
			sourceIDToName[id] = name
			idMin, idMax = min(idMin, id), max(idMax, id)
		}
	})

	if err != nil {
		return wrapError(err, ErrTypeListMigrationsOnDisk)
	}

	report.SourcesOnDisk = len(sourceIDToName)

	if report.SourcesOnDisk == 0 {
		return nil
	}

	connection, err := pgx.Connect(ctx, applier.DatabaseURL)
	if err != nil {
		return wrapError(err, ErrTypeDBConnect)
	}

	defer func() { _ = connection.Close(ctx) }()

	ddl := sqlres.DDL(applier.Project.Configuration.TableNames.AppliedMigrations)

	if err := execSimple(ctx, connection.PgConn(), ddl); err != nil {
		return wrapError(&ExecSQLError{Cause: err, SQL: ddl}, ErrTypeCreateDDL)
	}

	err = applier.ScanAppliedMigrations(ctx, connection, uint64(idMin), uint64(idMax), func(id uint64) {
		delete(sourceIDToName, source.MigrationID(id))
	})

	if err != nil {
		return err
	}

	report.PendingSources = len(sourceIDToName)

	if report.PendingSources == 0 {
		return nil
	}

	if err = applier.applyAllPending(ctx, connection, sourceIDToName, dupeIDToName); err != nil {
		return err
	}

	return nil
}

func (applier *Applier) ScanAppliedMigrations(
	ctx context.Context,
	conn *pgx.Conn,
	minID, maxID uint64,
	callback func(uint64),
) error {
	queryTemplate := "SELECT id FROM %s WHERE id >= $1 AND id <= $2"
	query := fmt.Sprintf(queryTemplate, applier.migrationsTableName())

	rows, err := conn.Query(ctx, query, minID, maxID)

	if err != nil {
		return wrapError(&ExecSQLError{Cause: err, SQL: query}, ErrTypeScanAppliedMigrations)
	}

	var id uint64
	_, err = pgx.ForEachRow(rows, []any{&id}, func() error {
		callback(id)

		return nil
	})

	if err != nil {
		return wrapError(&ExecSQLError{Cause: err, SQL: query}, ErrTypeScanAppliedMigrations)
	}

	return nil
}

func (applier *Applier) applyAllPending(
	ctx context.Context,
	conn *pgx.Conn,
	sourceIDToName map[source.MigrationID]string,
	dupeIDToName map[source.MigrationID]string,
) error {
	loader := source.Loader{MaxSQLFileSize: applier.MaxSQLFileSize}

	for _, id := range applier.getSortedMigrationIDs(sourceIDToName) {
		name := sourceIDToName[id]
		dupeName, hasDuplicate := dupeIDToName[id]

		if hasDuplicate {
			err := &ApplyMigrationError{
				Cause: &DuplicateMigrationError{Paths: []string{name, dupeName}},
				Name:  name,
			}

			return wrapError(err, ErrTypeApplyMigration)
		}

		sourceDir := filepath.Join(applier.Project.Dir, name)

		source, err := loader.LoadSource(sourceDir)

		if err != nil {
			return wrapError(&ApplyMigrationError{Cause: err, Name: name}, ErrTypeApplyMigration)
		}

		if duration, err := applier.applyMigration(ctx, conn, source.UpSQL, name); err != nil {
			return wrapError(&ApplyMigrationError{Cause: err, Name: name}, ErrTypeApplyMigration)
		} else if err := applier.registerMigration(ctx, conn, id, &source, duration); err != nil {
			return wrapError(err, ErrTypeRegisterMigration)
		}
	}

	return nil
}

func (applier *Applier) applyMigration(
	ctx context.Context,
	conn *pgx.Conn,
	sql string,
	name string,
) (time.Duration, error) {
	startTime := time.Now()

	log.Printf("Applying %q,please wait...", name)

	if isConnectionInTransaction(conn.PgConn()) {
		panic("the connection is not allowed to be in an active transaction")
	}

	if err := execSimple(ctx, conn.PgConn(), sql); err != nil {
		return 0, &ExecSQLError{Cause: err, SQL: sql}
	}

	if isConnectionInTransaction(conn.PgConn()) {
		err := execSimple(ctx, conn.PgConn(), "ROLLBACK;")

		return 0, &TransactionNotCommittedError{RollBackError: err}
	}

	duration := time.Since(startTime)
	durationStr := humanizeDuration(duration, "0ms")

	log.Printf("Applied  %q in %s", name, durationStr)

	return duration, nil
}

func (applier *Applier) registerMigration(
	ctx context.Context,
	conn *pgx.Conn,
	id source.MigrationID,
	source *source.Source,
	duration time.Duration,
) error {
	columns := []string{
		"id",
		"project",
		"name",
		"applied_at",
		"sql_up",
		"sql_down",
		"sql_up_sha256",
		"sql_down_sha256",
		"duration_ms",
		"rollback_blocked",
		"meta",
	}

	params := make([]string, len(columns))

	for i, column := range columns {
		params[i] = "@" + column
	}

	query := fmt.Sprintf(
		`INSERT INTO %s (%s) VALUES (%s);`,
		applier.migrationsTableName(),
		strings.Join(columns, ","),
		strings.Join(params, ","),
	)

	args := pgx.NamedArgs{
		"id":               id,
		"project":          applier.Project.Configuration.Name,
		"name":             source.Configuration.Name,
		"applied_at":       time.Now().UTC(),
		"sql_up":           source.UpSQL,
		"sql_down":         source.DownSQL,
		"sql_up_sha256":    Sha256ToHexStr(source.UpSQL),
		"sql_down_sha256":  Sha256ToHexStr(source.DownSQL),
		"duration_ms":      duration.Milliseconds(),
		"rollback_blocked": source.Configuration.Down.Block,
		"meta":             source.Configuration.Meta,
	}

	_, err := conn.Exec(ctx, query, args)

	if err != nil {
		return &ExecSQLError{Cause: err, SQL: query}
	}

	return nil
}

func (applier *Applier) getSortedMigrationIDs(sourceIDToName map[source.MigrationID]string) []source.MigrationID {
	sortedIDs := make([]source.MigrationID, 0, len(sourceIDToName))

	for id := range sourceIDToName {
		sortedIDs = append(sortedIDs, id)
	}

	slices.Sort(sortedIDs)

	return sortedIDs
}

func (applier *Applier) migrationsTableName() string {
	return applier.Project.Configuration.TableNames.AppliedMigrations
}
