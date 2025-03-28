package migrator

import (
	"context"
	"log"
	"math"
	"path/filepath"
	"slices"
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

	migrationsRepo *Migrations
}

func (applier *Applier) ApplyPending(ctx context.Context, report *Report) error {
	applier.migrationsRepo = &Migrations{TableName: applier.Project.Configuration.MigrationsTableName}

	sourceIDToName := make(map[uint64]string)
	dupeIDToName := make(map[uint64]string)
	idMin, idMax := uint64(math.MaxUint64), uint64(0)

	err := source.ScanAll(applier.Project.Dir, func(id uint64, name string) {
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

	ddl := sqlres.DDL(applier.Project.Configuration.MigrationsTableName)

	if err := execSimple(ctx, connection.PgConn(), ddl); err != nil {
		return wrapError(&ExecSQLError{Cause: err, SQL: ddl}, ErrTypeCreateDDL)
	}

	appliedIDs, err := applier.migrationsRepo.ScanApplied(ctx, connection, idMin, idMax)

	if err != nil {
		return wrapError(err, ErrTypeScanAppliedMigrations)
	}

	for _, id := range appliedIDs {
		delete(sourceIDToName, id)
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

func (applier *Applier) applyAllPending(
	ctx context.Context,
	conn *pgx.Conn,
	sourceIDToName map[uint64]string,
	dupeIDToName map[uint64]string,
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
	id uint64,
	source *source.Source,
	duration time.Duration,
) error {
	migration := &Migration{
		ID:              id,
		Name:            source.Configuration.Name,
		AppliedAt:       time.Now().UTC(),
		SQLUp:           source.UpSQL,
		SQLDown:         source.DownSQL,
		SQLUpSHA256:     Sha256ToHexStr(source.UpSQL),
		SQLDownSHA256:   Sha256ToHexStr(source.DownSQL),
		DurationMs:      duration.Milliseconds(),
		RollbackBlocked: source.Configuration.Down.Block,
		Meta:            source.Configuration.Meta,
	}

	return applier.migrationsRepo.Insert(ctx, conn, migration)
}

func (applier *Applier) getSortedMigrationIDs(sourceIDToName map[uint64]string) []uint64 {
	ids := mapKeysToSlice(sourceIDToName)

	slices.Sort(ids)

	return ids
}
