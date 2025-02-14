package migrator_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/servletcloud/Andmerada/internal/migrator"
	"github.com/servletcloud/Andmerada/internal/migrator/sqlres"
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

//nolint:paralleltest
func TestScanAppliedMigrations(t *testing.T) {
	ctx := context.Background()
	connectionURL := tests.StartEmbeddedPostgres(t)
	conn := tests.OpenPgConnection(t, connectionURL)
	dir := t.TempDir()

	_, err := conn.Exec(ctx, sqlres.DDL("applied_migrations"))
	require.NoError(t, err)

	applier := &migrator.Applier{
		Project: project.Project{
			Dir:           dir,
			Configuration: createProjectConfig(),
		},
		DatabaseURL: string(connectionURL),
	}

	scanAppliedMigrations := func(t *testing.T, minID, maxID uint64) []uint64 {
		t.Helper()

		result := make([]uint64, 0)

		err := applier.ScanAppliedMigrations(ctx, conn, minID, maxID, func(id uint64) {
			result = append(result, id)
		})

		require.NoError(t, err)

		return result
	}

	t.Run("empty table", func(t *testing.T) {
		actual := scanAppliedMigrations(t, 0, 99991225112129)
		require.Empty(t, actual)
	})

	t.Run("when there are applied migrations", func(t *testing.T) {
		insertDummyMigration(ctx, t, conn, 20241225112129)
		insertDummyMigration(ctx, t, conn, 20241225112130)
		insertDummyMigration(ctx, t, conn, 20241225112131)

		t.Run("filter covers the boundaries", func(t *testing.T) {
			actual := scanAppliedMigrations(t, 20241225112129, 20241225112131)

			assert.Len(t, actual, 3)
			assert.Contains(t, actual, uint64(20241225112129), uint64(20241225112130), uint64(20241225112131))
		})

		t.Run("filter includes the boundary values", func(t *testing.T) {
			actual := scanAppliedMigrations(t, 20241225112130, 20241225112130)

			assert.Len(t, actual, 1)
			assert.Contains(t, actual, uint64(20241225112130))
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

func insertDummyMigration(ctx context.Context, t *testing.T, conn *pgx.Conn, id uint64) {
	t.Helper()

	project := "test_scan_applied_migrations"
	name := "create users table"
	sqlUp := "create table users (id bigint primary key);"
	sqlUpSHA256 := "9473f4cfe827e5c29acffc4c80b8194aa3df919577fbf2f6b11df3d0f14cd907"
	durationMS := 10
	meta := make(map[string]struct{})

	query := `
		INSERT INTO applied_migrations (id, project, name, sql_up, sql_up_sha256, duration_ms, meta)
		VALUES ($1, $2, $3, $4, $5, $6, $7);
	`

	_, err := conn.Exec(ctx, query, id, project, name, sqlUp, sqlUpSHA256, durationMS, meta)
	require.NoError(t, err)
}
