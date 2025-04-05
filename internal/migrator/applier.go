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

type ApplyOptions struct {
	MaxSQLFileSize int64
	DatabaseURL    string
	Project        project.Project
}

type applier struct {
	maxSQLFileSize  int64
	databaseURL     string
	projectDir      string
	migrationsTable string

	report         *Report
	migrationsRepo *Migrations
	connection     *pgx.Conn
}

func ApplyPending(ctx context.Context, options ApplyOptions, report *Report) error {
	projectConfiguration := options.Project.Configuration
	migrationsTable := projectConfiguration.MigrationsTableName

	report.PendingCount = 0

	applier := &applier{
		maxSQLFileSize:  options.MaxSQLFileSize,
		databaseURL:     options.DatabaseURL,
		projectDir:      options.Project.Dir,
		report:          report,
		migrationsTable: migrationsTable,
		migrationsRepo:  &Migrations{TableName: migrationsTable},
		connection:      nil,
	}

	defer applier.close(ctx)

	return applier.applyPending(ctx)
}

func (applier *applier) applyPending(ctx context.Context) error {
	sources, err := source.ScanAll(applier.projectDir)
	if err != nil {
		return wrapError(err, ErrTypeListMigrationsOnDisk)
	}

	if len(sources) == 0 {
		return nil
	}

	if err := applier.connect(ctx); err != nil {
		return wrapError(err, ErrTypeDBConnect)
	}

	if err := applier.migrationsRepo.RunDDL(ctx, applier.connection); err != nil {
		return wrapError(err, ErrTypeCreateDDL)
	}

	appliedIDs, err := applier.scanAppliedMigrations(ctx, sources)
	if err != nil {
		return wrapError(err, ErrTypeScanAppliedMigrations)
	}

	deleteAllKeys(sources, appliedIDs)

	if len(sources) == 0 {
		return nil
	}

	applier.report.PendingCount = len(sources)

	return applier.applyAll(ctx, sources)
}

func (applier *applier) scanAppliedMigrations(ctx context.Context, sources map[uint64]string) ([]uint64, error) {
	idMin, idMax := uint64(math.MaxUint64), uint64(0)

	for id := range sources {
		idMin = min(idMin, id)
		idMax = max(idMax, id)
	}

	ids, err := applier.migrationsRepo.ScanApplied(ctx, applier.connection, idMin, idMax)

	if err != nil {
		return nil, err
	}

	return ids, nil
}

func (applier *applier) applyAll(ctx context.Context, sources map[uint64]string) error {
	loader := source.Loader{MaxSQLFileSize: applier.maxSQLFileSize}
	ids := mapKeysToSlice(sources)

	slices.Sort(ids)

	for _, id := range ids {
		name := sources[id]

		sourceDir := filepath.Join(applier.projectDir, name)

		source, err := loader.LoadSource(sourceDir)

		if err != nil {
			return wrapError(&ApplyMigrationError{Cause: err, Name: name}, ErrTypeApplyMigration)
		}

		if duration, err := applier.applyMigration(ctx, source.UpSQL, name); err != nil {
			return wrapError(&ApplyMigrationError{Cause: err, Name: name}, ErrTypeApplyMigration)
		} else if err := applier.registerMigration(ctx, id, &source, duration); err != nil {
			return wrapError(err, ErrTypeRegisterMigration)
		}
	}

	return nil
}

func (applier *applier) applyMigration(ctx context.Context, sql string, name string) (time.Duration, error) {
	startTime := time.Now()

	log.Printf("Applying %q,please wait...", name)

	if err := applier.executeMigrationSQL(ctx, sql); err != nil {
		return 0, err
	}

	duration := time.Since(startTime)
	durationStr := humanizeDuration(duration, "0ms")

	log.Printf("Applied  %q in %s", name, durationStr)

	return duration, nil
}

func (applier *applier) executeMigrationSQL(ctx context.Context, sql string) error {
	pgConn := applier.connection.PgConn()

	if isConnectionInTransaction(pgConn) {
		panic("the connection is not allowed to be in an active transaction")
	}

	if err := execSimple(ctx, pgConn, sql); err != nil {
		return &ExecSQLError{Cause: err, SQL: sql}
	}

	if isConnectionInTransaction(pgConn) {
		err := execSimple(ctx, pgConn, "ROLLBACK;")

		return &TransactionNotCommittedError{RollBackError: err}
	}

	return nil
}

func (applier *applier) registerMigration(
	ctx context.Context,
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

	return applier.migrationsRepo.Insert(ctx, applier.connection, migration)
}

func (applier *applier) connect(ctx context.Context) error {
	applier.connection = nil

	connection, err := pgx.Connect(ctx, applier.databaseURL)

	if err != nil {
		return err //nolint:wrapcheck
	}

	applier.connection = connection

	return nil
}

func (applier *applier) close(ctx context.Context) error {
	if applier.connection == nil {
		return nil
	}

	err := applier.connection.Close(ctx)

	if err != nil {
		log.Println("Failed to close the database connection:", err)
	}

	return err //nolint:wrapcheck
}
