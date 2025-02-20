package migrator

import (
	"context"
	"fmt"
	"math"

	"github.com/jackc/pgx/v5"
	"github.com/servletcloud/Andmerada/internal/migrator/sqlres"
	"github.com/servletcloud/Andmerada/internal/project"
	"github.com/servletcloud/Andmerada/internal/source"
)

type Report struct {
	SourcesOnDisk int
}

type Applier struct {
	DatabaseURL string
	Project     project.Project
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

	if len(sourceIDToName) == 0 {
		return nil
	}

	connection, err := pgx.Connect(ctx, applier.DatabaseURL)
	if err != nil {
		return wrapError(err, ErrTypeDBConnect)
	}

	defer func() { _ = connection.Close(ctx) }()

	ddl := sqlres.DDL(applier.Project.Configuration.TableNames.AppliedMigrations)

	if err := execSimple(ctx, connection.PgConn(), ddl); err != nil {
		return wrapErrorWithSQL(err, ErrTypeCreateDDL, ddl)
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
		return wrapErrorWithSQL(err, ErrTypeScanAppliedMigrations, query)
	}

	var id uint64
	_, err = pgx.ForEachRow(rows, []any{&id}, func() error {
		callback(id)

		return nil
	})

	if err != nil {
		return wrapErrorWithSQL(err, ErrTypeScanAppliedMigrations, query)
	}

	return nil
}

func (applier *Applier) migrationsTableName() string {
	return applier.Project.Configuration.TableNames.AppliedMigrations
}
