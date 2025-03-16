package migrator_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/servletcloud/Andmerada/internal/migrator"
	"github.com/servletcloud/Andmerada/internal/migrator/sqlres"
	"github.com/servletcloud/Andmerada/internal/project"
	"github.com/servletcloud/Andmerada/internal/resources"
	"github.com/servletcloud/Andmerada/internal/source"
	"github.com/servletcloud/Andmerada/internal/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//nolint:paralleltest,funlen
func TestApplyPending(t *testing.T) {
	ctx := context.Background()
	connectionURL := tests.StartEmbeddedPostgres(t)
	conn := tests.OpenPgConnection(t, connectionURL)
	dir := t.TempDir()
	report := migrator.Report{} //nolint:exhaustruct

	applier := &migrator.Applier{
		MaxSQLFileSize: 1024,
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

	t.Run("There are no migrations", func(t *testing.T) {
		t.Run("It does not create a migrations table", func(t *testing.T) {
			tests.AssertPgTableNotExist(ctx, t, conn, "applied_migrations")

			mustApplyPending(t)

			tests.AssertPgTableNotExist(ctx, t, conn, "applied_migrations")
		})

		t.Run("Report has zero migrations", func(t *testing.T) {
			mustApplyPending(t)
			assert.Equal(t, 0, report.SourcesOnDisk)
		})
	})

	t.Run("When there are migrations on the disk", func(t *testing.T) {
		createMigration := func(t *testing.T, title string, timestamp string, upSQL string) source.CreateSourceResult {
			t.Helper()

			source := createSource(t, dir, title, timestamp)
			writeUpSQL(t, source.FullPath, upSQL)

			return source
		}

		createMigration(t, "Cr table", "20241225112129", "CREATE TABLE users (id SERIAL PRIMARY KEY, name TEXT NOT NULL);")
		createMigration(t, "Add column", "20241228143752", "ALTER TABLE users ADD COLUMN age INTEGER NOT NULL DEFAULT 0;")
		createMigration(t, "Rename column", "20241229143752", "ALTER TABLE users RENAME COLUMN age TO years;")

		t.Run("It applies the migrations in order", func(t *testing.T) {
			tests.AssertPgTableNotExist(ctx, t, conn, "applied_migrations")
			tests.AssertPgTableNotExist(ctx, t, conn, "users")

			mustApplyPending(t)

			tests.AssertPgTableExist(ctx, t, conn, "applied_migrations")
			tests.AssertPgTableExist(ctx, t, conn, "users")

			assert.Equal(t, 3, report.SourcesOnDisk)
			assert.Equal(t, 3, report.PendingSources)
		})

		t.Run("Insert into a newly create table succeeds", func(t *testing.T) {
			_, err := conn.Exec(ctx, "INSERT INTO users (id, name, years) VALUES (1, 'John Doe', 30)")
			require.NoError(t, err)
		})

		t.Run("It is idempotent, does not apply the migrations again", func(t *testing.T) {
			tests.AssertPgTableExist(ctx, t, conn, "applied_migrations")
			tests.AssertPgTableExist(ctx, t, conn, "users")

			mustApplyPending(t)

			assert.Equal(t, 3, report.SourcesOnDisk)
			assert.Equal(t, 0, report.PendingSources)
		})

		t.Run("Fails to apply a not committed migration", func(t *testing.T) {
			sql := "BEGIN; CREATE TABLE sessions (id SERIAL PRIMARY KEY, user_id INTEGER);"
			createMigration(t, "Create session table", "20250101101523", sql)

			err := applier.ApplyPending(ctx, &report)

			var notCommittedErr *migrator.TransactionNotCommittedError

			require.ErrorAs(t, err, &notCommittedErr)
			tests.AssertPgTableNotExist(ctx, t, conn, "sessions")
		})

		t.Run("Failed to apply a duplicate migration", func(t *testing.T) {
			sql := "CREATE TABLE sessions (id SERIAL PRIMARY KEY, user_id INTEGER);"
			createMigration(t, "Create session table", "20250101101523", sql)

			result := createMigration(t, "Some dummy migtation", "20250101101524", "CREATE TABLE dummy(id INTEGER);")
			tests.MkDir(t, result.FullPath+"_dupe")

			err := applier.ApplyPending(ctx, &report)

			tests.AssertPgTableExist(ctx, t, conn, "sessions")
			tests.AssertPgTableNotExist(ctx, t, conn, "dummy")

			var dupErr *migrator.DuplicateMigrationError

			require.ErrorAs(t, err, &dupErr)

			assert.Equal(t, []string{result.BaseDir, result.BaseDir + "_dupe"}, dupErr.Paths)
		})

		t.Run("Populates columns of applied_migrations table", func(t *testing.T) {
			createMigration(t, "Dummy migration", "20250109025508", "SELECT 1;")

			err := applier.ApplyPending(ctx, &report)
			require.NoError(t, err)

			t.Run("Populates id,project,name,applied_at columns", func(t *testing.T) {
				query := "SELECT id, project, name, applied_at FROM applied_migrations WHERE id=20250109025508"
				row := conn.QueryRow(ctx, query)

				var (
					id        uint64
					project   string
					name      string
					appliedAt time.Time
				)

				err := row.Scan(&id, &project, &name, &appliedAt)
				require.NoError(t, err)

				assert.Equal(t, uint64(20250109025508), id)
				assert.Equal(t, "migrator_test", project)
				assert.Equal(t, "Dummy migration", name)
				assert.False(t, appliedAt.IsZero())
			})

			t.Run("Populates sql and hashes columns", func(t *testing.T) {
				query := "SELECT sql_up,sql_down,sql_up_sha256,sql_down_sha256 FROM applied_migrations WHERE id=20250109025508"
				row := conn.QueryRow(ctx, query)

				var sqlUp, sqlDown, sqlUpSHA256, sqlDownSHA256 string

				err := row.Scan(&sqlUp, &sqlDown, &sqlUpSHA256, &sqlDownSHA256)
				require.NoError(t, err)

				assert.Equal(t, "SELECT 1;", sqlUp)
				assert.Equal(t, resources.TemplateDownSQL(), sqlDown)
				assert.Equal(t, migrator.Sha256ToHexStr(sqlUp), sqlUpSHA256)
				assert.Equal(t, migrator.Sha256ToHexStr(sqlDown), sqlDownSHA256)
			})

			t.Run("Populates duration_ms,rollback_blocked,meta columns", func(t *testing.T) {
				query := "SELECT duration_ms,rollback_blocked,meta FROM applied_migrations WHERE id=20250109025508"
				row := conn.QueryRow(ctx, query)

				var (
					durationMs      int64
					rollbackBlocked bool
					meta            map[string]any
				)

				err := row.Scan(&durationMs, &rollbackBlocked, &meta)
				require.NoError(t, err)

				assert.GreaterOrEqual(t, durationMs, int64(0))
				assert.False(t, rollbackBlocked)
				assert.Equal(t, map[string]any{"description": "Full description of the migration"}, meta)
			})
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
		MaxSQLFileSize: 1024,
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

func createSource(t *testing.T, dir string, title string, timestamp string) source.CreateSourceResult {
	t.Helper()

	timeParsed, err := time.Parse("20060102150405", timestamp)
	require.NoError(t, err)

	result, err := source.Create(dir, title, timeParsed)
	require.NoError(t, err)

	t.Cleanup(func() {
		err := os.RemoveAll(result.FullPath)
		require.NoError(t, err)
	})

	return result
}

func writeUpSQL(t *testing.T, dir string, content string) {
	t.Helper()

	path := filepath.Join(dir, source.UpSQLFilename)
	err := os.WriteFile(path, []byte(content), 0600)

	require.NoError(t, err)
}
