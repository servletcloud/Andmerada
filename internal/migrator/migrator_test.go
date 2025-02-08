package migrator_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/servletcloud/Andmerada/internal/migrator"
	"github.com/servletcloud/Andmerada/internal/project"
	"github.com/servletcloud/Andmerada/internal/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//nolint:paralleltest
func TestApplyPending(t *testing.T) {
	ctx := context.Background()
	connectionURL := tests.StartEmbeddedPostgres(t)
	conn := tests.OpenPgConnection(t, connectionURL)
	dir := t.TempDir()
	report := migrator.Report{} //nolint:exhaustruct

	applier := &migrator.Applier{
		Project: project.Project{
			Dir:           dir,
			Configuration: createProjectConfig(),
		},
		DatabaseURL: string(connectionURL),
	}

	mustApplyPending := func(t *testing.T) {
		t.Helper()

		err := applier.ApplyPending(ctx, &report)
		require.NoError(t, err)
	}

	assertMigrationsTableExists := func(t *testing.T, expected bool) {
		t.Helper()

		actual := tests.IsPgTableExist(ctx, t, conn, "applied_migrations")

		require.Equal(t, expected, actual, "existence of 'applied_migrations' table")
	}

	t.Run("Does not run DDL if there are no migrations", func(t *testing.T) {
		assertMigrationsTableExists(t, false)
		mustApplyPending(t)
		assertMigrationsTableExists(t, false)
	})

	t.Run("Creates 'applied_migrations' table", func(t *testing.T) {
		tests.MkDir(t, filepath.Join(dir, "20241225112129_create_users_table"))
		assertMigrationsTableExists(t, false)
		mustApplyPending(t)
		assertMigrationsTableExists(t, true)
	})

	t.Run("Running DDL is idempotent", func(t *testing.T) {
		tests.MkDir(t, filepath.Join(dir, "20241225112129_create_users_table"))
		assertMigrationsTableExists(t, true)
		mustApplyPending(t)
	})

	t.Run("report.SourcesOnDisk", func(_ *testing.T) {
		t.Run("returns 0 if there are no migrations on the disk", func(t *testing.T) {
			mustApplyPending(t)
			assert.Equal(t, 0, report.SourcesOnDisk)
		})

		t.Run("returns 1 if there are a migration on the disk", func(t *testing.T) {
			tests.MkDir(t, filepath.Join(dir, "20241225112129_create_users_table"))
			mustApplyPending(t)
			assert.Equal(t, 1, report.SourcesOnDisk)
		})
	})
}

func createProjectConfig() project.Configuration {
	return project.Configuration{
		Name: "migrator_test",
		TableNames: struct {
			AppliedMigrations string `yaml:"applied_migrations"`
		}{
			AppliedMigrations: "applied_migrations",
		},
	}
}
