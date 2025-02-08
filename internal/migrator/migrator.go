package migrator

import (
	"context"
	"fmt"

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

	err := source.ScanAll(applier.Project.Dir, func(id source.MigrationID, name string) {
		_, found := sourceIDToName[id]
		if found {
			dupeIDToName[id] = name
		} else {
			sourceIDToName[id] = name
		}
	})

	if err != nil {
		return fmt.Errorf("failed to scan migration files on disk: %w", err)
	}

	report.SourcesOnDisk = len(sourceIDToName)

	if len(sourceIDToName) == 0 {
		return nil
	}

	connection, err := pgx.Connect(ctx, applier.DatabaseURL)
	if err != nil {
		return &PostgresConnectError{cause: err}
	}

	defer func() { _ = connection.Close(ctx) }()

	ddl := sqlres.DDL(applier.Project.Configuration.TableNames.AppliedMigrations)

	if err := execSimple(ctx, connection.PgConn(), ddl); err != nil {
		return &CreateDDLFailedError{cause: err, SQL: ddl}
	}

	return nil
}
