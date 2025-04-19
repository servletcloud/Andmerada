package migrator

import (
	"cmp"
	"context"
	"iter"
	"log"
	"maps"
	"path/filepath"
	"slices"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/servletcloud/Andmerada/internal/project"
	"github.com/servletcloud/Andmerada/internal/source"
)

type Report struct {
	PendingCount int
}

type ApplyOptions struct {
	MaxSQLFileSize    int64
	DatabaseURL       string
	Project           project.Project
	Limit             int
	DryRun            bool
	SkipPreValidation bool
}

type applier struct {
	maxSQLFileSize    int64
	databaseURL       string
	projectDir        string
	migrationsTable   string
	limit             int
	dryRun            bool
	skipPreValidation bool

	report         *Report
	migrationsRepo *Migrations
	loader         source.Loader
	connection     *pgx.Conn
}

type sourceRef struct {
	id   source.ID
	name string
}

const (
	NoLimit = 0
)

func ApplyPending(ctx context.Context, options ApplyOptions, report *Report) error {
	projectConfiguration := options.Project.Configuration
	migrationsTable := projectConfiguration.MigrationsTableName

	report.PendingCount = 0

	applier := &applier{
		maxSQLFileSize:    options.MaxSQLFileSize,
		databaseURL:       options.DatabaseURL,
		projectDir:        options.Project.Dir,
		limit:             options.Limit,
		dryRun:            options.DryRun,
		skipPreValidation: options.SkipPreValidation,
		report:            report,
		migrationsTable:   migrationsTable,
		migrationsRepo:    &Migrations{TableName: migrationsTable},
		loader:            source.Loader{MaxSQLFileSize: options.MaxSQLFileSize},
		connection:        nil,
	}

	defer applier.close(ctx)

	return applier.applyPending(ctx)
}

func (applier *applier) applyPending(ctx context.Context) error {
	sourceIDToName, err := source.ScanAll(applier.projectDir)
	if err != nil {
		return wrapError(err, ErrTypeListMigrationsOnDisk)
	}

	if len(sourceIDToName) == 0 {
		return nil
	}

	if err := applier.connect(ctx); err != nil {
		return wrapError(err, ErrTypeDBConnect)
	}

	appliedIDs, err := applier.scanAppliedMigrations(ctx, maps.Keys(sourceIDToName))
	if err != nil {
		needsToRunDDL := isPgErrorOfCode(err, pgerrcode.UndefinedTable)

		if !needsToRunDDL {
			return wrapError(err, ErrTypeScanAppliedMigrations)
		}

		if err := applier.runDDL(ctx); err != nil {
			return wrapError(err, ErrTypeCreateDDL)
		}
	}

	for _, appliedID := range appliedIDs {
		delete(sourceIDToName, appliedID)
	}

	sourceRefs := applier.toSortedSourceRefs(sourceIDToName)
	applier.report.PendingCount = len(sourceRefs)

	if err := applier.preValidateSources(sourceRefs); err != nil {
		return wrapError(err, ErrTypePreValidateSources)
	}

	return applier.applyAll(ctx, sourceRefs)
}

func (applier *applier) runDDL(ctx context.Context) error {
	if applier.dryRun {
		return nil
	}

	return applier.migrationsRepo.RunDDL(ctx, applier.connection)
}

func (applier *applier) scanAppliedMigrations(
	ctx context.Context,
	availableIDs iter.Seq[source.ID],
) ([]source.ID, error) {
	idMin, idMax := source.MinMigrationID, source.MaxMigrationID

	for id := range availableIDs {
		idMin = min(idMin, id)
		idMax = max(idMax, id)
	}

	return applier.migrationsRepo.ScanApplied(ctx, applier.connection, idMin, idMax)
}

func (applier *applier) toSortedSourceRefs(sources map[source.ID]string) []sourceRef {
	result := make([]sourceRef, 0, len(sources))

	for id, name := range sources {
		result = append(result, sourceRef{id, name})
	}

	slices.SortFunc(result, func(a, b sourceRef) int {
		return cmp.Compare(a.id, b.id)
	})

	if applier.limit == NoLimit {
		return result
	}

	upperBound := min(len(result), applier.limit)

	return result[:upperBound]
}

func (applier *applier) preValidateSources(sourceRefs []sourceRef) error {
	if applier.skipPreValidation {
		return nil
	}

	source := source.Source{} //nolint:exhaustruct

	for _, ref := range sourceRefs {
		if err := applier.loader.ValidateSource(filepath.Join(applier.projectDir, ref.name), &source); err != nil {
			return &LoadSourceError{Cause: err, Name: ref.name}
		}
	}

	return nil
}

func (applier *applier) applyAll(ctx context.Context, sourceRefs []sourceRef) error {
	source := source.Source{} //nolint:exhaustruct

	for _, ref := range sourceRefs {
		name := ref.name

		if err := applier.loadSource(ref, &source); err != nil {
			return wrapError(err, ErrTypeLoadMigration)
		}

		if duration, err := applier.applyMigration(ctx, source.UpSQL, ref); err != nil {
			return wrapError(&ApplyMigrationError{Cause: err, Name: name}, ErrTypeApplyMigration)
		} else if err := applier.registerMigration(ctx, ref, &source, duration); err != nil {
			return wrapError(err, ErrTypeRegisterMigration)
		}
	}

	return nil
}

func (applier *applier) loadSource(ref sourceRef, out *source.Source) error {
	dir := filepath.Join(applier.projectDir, ref.name)

	if err := applier.loader.LoadSource(dir, out); err != nil {
		return &LoadSourceError{Cause: err, Name: ref.name}
	}

	return nil
}

func (applier *applier) applyMigration(ctx context.Context, sql string, ref sourceRef) (time.Duration, error) {
	startTime := time.Now()

	log.Printf("Applying %q, please wait...", ref.name)

	if err := applier.executeMigrationSQL(ctx, sql); err != nil {
		return 0, err
	}

	duration := time.Since(startTime)
	durationStr := humanizeDuration(duration, "0ms")

	log.Printf("Applied  %q in %s", ref.name, durationStr)

	return duration, nil
}

func (applier *applier) executeMigrationSQL(ctx context.Context, sql string) error {
	pgConn := applier.connection.PgConn()

	if isConnectionInTransaction(pgConn) {
		panic("the connection is not allowed to be in an active transaction")
	}

	if !applier.dryRun {
		if err := execSimple(ctx, pgConn, sql); err != nil {
			return &ExecSQLError{Cause: err, SQL: sql}
		}
	}

	if isConnectionInTransaction(pgConn) {
		err := execSimple(ctx, pgConn, "ROLLBACK;")

		return &TransactionNotCommittedError{RollBackError: err}
	}

	return nil
}

func (applier *applier) registerMigration(
	ctx context.Context,
	ref sourceRef,
	source *source.Source,
	duration time.Duration,
) error {
	if applier.dryRun {
		return nil
	}

	migration := &Migration{
		ID:              ref.id,
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
