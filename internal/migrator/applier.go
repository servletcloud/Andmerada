package migrator

import (
	"context"
	"log"
	"math"
	"path/filepath"
	"slices"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/servletcloud/Andmerada/internal/project"
	"github.com/servletcloud/Andmerada/internal/source"
)

type Report struct {
	PendingCount int
}

type Applier struct {
	MaxSQLFileSize int64
	DatabaseURL    string
	Project        project.Project
}

func (applier *Applier) ApplyPending(ctx context.Context, report *Report) error {
	sources, err := applier.scanAvailableMigrations()
	if err != nil {
		return err
	}

	if len(sources) == 0 {
		return nil
	}

	connection, err := pgx.Connect(ctx, applier.DatabaseURL)
	if err != nil {
		return wrapError(err, ErrTypeDBConnect)
	}

	defer func() { _ = connection.Close(ctx) }()

	if err := applier.runDDL(ctx, connection); err != nil {
		return err
	}

	appliedIDs, err := applier.scanAppliedMigrations(ctx, connection, sources)
	if err != nil {
		return err
	}

	deleteAllKeys(sources, appliedIDs)

	report.PendingCount = len(sources)

	if report.PendingCount == 0 {
		return nil
	}

	return applier.applyAll(ctx, connection, sources)
}

func (applier *Applier) scanAvailableMigrations() (map[uint64]string, error) {
	sources, err := source.ScanAll(applier.Project.Dir)
	if err != nil {
		return nil, wrapError(err, ErrTypeListMigrationsOnDisk)
	}

	return sources, nil
}

func (applier *Applier) runDDL(ctx context.Context, conn *pgx.Conn) error {
	migrationsRepo := Migrations{TableName: applier.Project.Configuration.MigrationsTableName}

	if err := migrationsRepo.RunDDL(ctx, conn); err != nil {
		return wrapError(err, ErrTypeCreateDDL)
	}

	return nil
}

func (applier *Applier) scanAppliedMigrations(
	ctx context.Context,
	conn *pgx.Conn,
	sources map[uint64]string,
) ([]uint64, error) {
	idMin, idMax := uint64(math.MaxUint64), uint64(0)

	for id := range sources {
		idMin = min(idMin, id)
		idMax = max(idMax, id)
	}

	migrationsRepo := &Migrations{TableName: applier.Project.Configuration.MigrationsTableName}
	ids, err := migrationsRepo.ScanApplied(ctx, conn, idMin, idMax)

	if err != nil {
		return nil, wrapError(err, ErrTypeScanAppliedMigrations)
	}

	return ids, nil
}

func (applier *Applier) applyAll(
	ctx context.Context,
	conn *pgx.Conn,
	sources map[uint64]string,
) error {
	loader := source.Loader{MaxSQLFileSize: applier.MaxSQLFileSize}

	for _, id := range applier.getSortedMigrationIDs(sources) {
		name := sources[id]

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

	migrationsRepo := &Migrations{TableName: applier.Project.Configuration.MigrationsTableName}

	return migrationsRepo.Insert(ctx, conn, migration)
}

func (applier *Applier) getSortedMigrationIDs(sourceIDToName map[uint64]string) []uint64 {
	ids := mapKeysToSlice(sourceIDToName)

	slices.Sort(ids)

	return ids
}
