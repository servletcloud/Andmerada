package migrator_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/servletcloud/Andmerada/internal/migrator"
	"github.com/servletcloud/Andmerada/internal/project"
	"github.com/servletcloud/Andmerada/internal/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//nolint:paralleltest
func TestExecSimple(t *testing.T) {
	ctx := context.Background()
	connectionURL := tests.StartEmbeddedPostgres(t)
	conn := tests.OpenPgConnection(t, connectionURL)

	applier := &migrator.Applier{
		Project: project.Project{
			Dir:           t.TempDir(),
			Configuration: createProjectConfig(),
		},
		DatabaseURL: string(connectionURL),
	}

	t.Run("Creates 'applied_migrations' table", func(t *testing.T) {
		_, err := conn.Exec(ctx, "select count(id) from applied_migrations;")

		var pgError *pgconn.PgError

		require.ErrorAs(t, err, &pgError)
		require.Equal(t, "42P01", pgError.Code)

		err = applier.ApplyPending(ctx)
		require.NoError(t, err)

		_, err = conn.Exec(ctx, "select count(id) from applied_migrations;")
		require.NoError(t, err)
	})

	t.Run("Running DDL is idempotent", func(t *testing.T) {
		assert.NoError(t, applier.ApplyPending(ctx))
		assert.NoError(t, applier.ApplyPending(ctx))
	})
}

func createProjectConfig() project.Configuration {
	return project.Configuration{
		Name: "migrator_test",
		TableNames: struct {
			AppliedMigrations string `yaml:"applied_migrations"` //nolint:tagliatelle
		}{
			AppliedMigrations: "applied_migrations",
		},
	}
}
