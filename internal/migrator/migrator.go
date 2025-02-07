package migrator

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/servletcloud/Andmerada/internal/migrator/sqlres"
	"github.com/servletcloud/Andmerada/internal/project"
)

type Applier struct {
	DatabaseURL string
	Project     project.Project
}

func (applier *Applier) ApplyPending(ctx context.Context) error {
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
