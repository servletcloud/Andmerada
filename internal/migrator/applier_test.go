package migrator_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/servletcloud/Andmerada/internal/migrator"
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
	report := migrator.Report{PendingCount: 0}

	options := migrator.ApplyOptions{
		MaxSQLFileSize: 1024,
		Project: project.Project{
			Dir:           dir,
			Configuration: createProjectConfig(),
		},
		DatabaseURL:       string(connectionURL),
		SkipPreValidation: false,
		DryRun:            false,
	}

	mustApplyPending := func(t *testing.T) {
		t.Helper()

		err := migrator.ApplyPending(ctx, options, &report)
		require.NoError(t, err)
	}

	t.Run("There are no migrations", func(t *testing.T) {
		t.Run("It does not create a migrations table", func(t *testing.T) {
			tests.AssertPgTableNotExist(ctx, t, conn, "migrations")

			mustApplyPending(t)

			tests.AssertPgTableNotExist(ctx, t, conn, "migrations")
		})

		t.Run("Report has zero migrations", func(t *testing.T) {
			mustApplyPending(t)
			assert.Equal(t, 0, report.PendingCount)
		})
	})

	t.Run("When there are migrations on the disk", func(t *testing.T) {
		createMigration := func(t *testing.T, title string, timestamp string, upSQL string) source.CreateSourceResult {
			t.Helper()

			source := tests.CreateSource(t, dir, title, timestamp)
			writeUpSQL(t, source.FullPath, upSQL)

			return source
		}

		createMigration(t, "Cr table", "20241225112129", "CREATE TABLE users (id SERIAL PRIMARY KEY, name TEXT NOT NULL);")
		createMigration(t, "Add column", "20241228143752", "ALTER TABLE users ADD COLUMN age INTEGER NOT NULL DEFAULT 0;")
		createMigration(t, "Rename column", "20241229143752", "ALTER TABLE users RENAME COLUMN age TO years;")

		t.Run("In dry run mode, it does not apply the migrations", func(t *testing.T) {
			tests.AssertPgTableNotExist(ctx, t, conn, "migrations")
			tests.AssertPgTableNotExist(ctx, t, conn, "users")

			optionsCopy := options
			optionsCopy.DryRun = true

			err := migrator.ApplyPending(ctx, optionsCopy, &report)
			require.NoError(t, err)

			tests.AssertPgTableNotExist(ctx, t, conn, "migrations")
			tests.AssertPgTableNotExist(ctx, t, conn, "users")
		})

		t.Run("It applies the migrations in order", func(t *testing.T) {
			tests.AssertPgTableNotExist(ctx, t, conn, "migrations")
			tests.AssertPgTableNotExist(ctx, t, conn, "users")

			mustApplyPending(t)

			tests.AssertPgTableExist(ctx, t, conn, "migrations")
			tests.AssertPgTableExist(ctx, t, conn, "users")

			assert.Equal(t, 3, report.PendingCount)
		})

		t.Run("Insert into a newly create table succeeds", func(t *testing.T) {
			_, err := conn.Exec(ctx, "INSERT INTO users (id, name, years) VALUES (1, 'John Doe', 30)")
			require.NoError(t, err)
		})

		t.Run("It is idempotent, does not apply the migrations again", func(t *testing.T) {
			tests.AssertPgTableExist(ctx, t, conn, "migrations")
			tests.AssertPgTableExist(ctx, t, conn, "users")

			mustApplyPending(t)

			assert.Equal(t, 0, report.PendingCount)
		})

		t.Run("Fails to apply a not committed migration", func(t *testing.T) {
			sql := "BEGIN; CREATE TABLE sessions (id SERIAL PRIMARY KEY, user_id INTEGER);"
			createMigration(t, "Create session table", "20250101101523", sql)

			err := migrator.ApplyPending(ctx, options, &report)

			var notCommittedErr *migrator.TransactionNotCommittedError

			require.ErrorAs(t, err, &notCommittedErr)
			tests.AssertPgTableNotExist(ctx, t, conn, "sessions")
		})

		t.Run("Failed all migrations because of a duplicate migration", func(t *testing.T) {
			sql := "CREATE TABLE sessions (id SERIAL PRIMARY KEY, user_id INTEGER);"
			createMigration(t, "Create session table", "20250101101523", sql)

			result := createMigration(t, "Some dummy migtation", "20250101101524", "CREATE TABLE dummy(id INTEGER);")
			tests.MkDir(t, result.FullPath+"_dupe")

			err := migrator.ApplyPending(ctx, options, &report)

			tests.AssertPgTableNotExist(ctx, t, conn, "sessions")
			tests.AssertPgTableNotExist(ctx, t, conn, "dummy")

			var dupErr *source.DuplicateSourceError

			require.ErrorAs(t, err, &dupErr)

			assert.Equal(t, []string{result.BaseDir, result.BaseDir + "_dupe"}, dupErr.Paths)
		})

		t.Run("Populates columns of migrations table", func(t *testing.T) {
			createMigration(t, "Dummy migration", "20250109025508", "SELECT 1;")

			err := migrator.ApplyPending(ctx, options, &report)
			require.NoError(t, err)

			t.Run("Populates id,name,applied_at columns", func(t *testing.T) {
				query := "SELECT id, name, applied_at FROM migrations WHERE id=20250109025508"
				row := conn.QueryRow(ctx, query)

				var (
					id        uint64
					name      string
					appliedAt time.Time
				)

				err := row.Scan(&id, &name, &appliedAt)
				require.NoError(t, err)

				assert.Equal(t, uint64(20250109025508), id)
				assert.Equal(t, "Dummy migration", name)
				assert.False(t, appliedAt.IsZero())
			})

			t.Run("Populates sql and hashes columns", func(t *testing.T) {
				query := "SELECT sql_up,sql_down,sql_up_sha256,sql_down_sha256 FROM migrations WHERE id=20250109025508"
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
				query := "SELECT duration_ms,rollback_blocked,meta FROM migrations WHERE id=20250109025508"
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

	t.Run("Pre-validation", func(t *testing.T) {
		dir := t.TempDir()
		preValidateOptions := options
		preValidateOptions.Project.Dir = dir

		source1 := tests.CreateSource(t, dir, "Valid migration 1", "20250122110947")
		source2 := tests.CreateSource(t, dir, "Valid migration 2", "20250318142235")

		err := os.Remove(filepath.Join(source2.FullPath, source.UpSQLFilename))
		require.NoError(t, err)

		t.Run("skip-pre-validation flag is false", func(t *testing.T) {
			writeUpSQL(t, source1.FullPath, "CREATE TABLE pre_validation_1 (id INTEGER);")

			optionsCopy := preValidateOptions
			optionsCopy.SkipPreValidation = false

			err = migrator.ApplyPending(ctx, optionsCopy, &report)

			var applierErr *migrator.MigrateError

			require.ErrorAs(t, err, &applierErr)
			assert.Equal(t, migrator.ErrTypePreValidateSources, applierErr.ErrType)

			var loadSourceErr *migrator.LoadSourceError

			require.ErrorAs(t, err, &loadSourceErr)
			assert.Equal(t, source2.BaseDir, loadSourceErr.Name)

			tests.AssertPgTableNotExist(ctx, t, conn, "pre_validation_1")
		})

		t.Run("skip-pre-validation flag is true", func(t *testing.T) {
			writeUpSQL(t, source1.FullPath, "CREATE TABLE pre_validation_2 (id INTEGER);")

			optionsCopy := preValidateOptions
			optionsCopy.SkipPreValidation = true

			err = migrator.ApplyPending(ctx, optionsCopy, &report)

			var applierErr *migrator.MigrateError

			require.ErrorAs(t, err, &applierErr)
			assert.Equal(t, migrator.ErrTypeLoadMigration, applierErr.ErrType)

			var loadErr *migrator.LoadSourceError

			require.ErrorAs(t, err, &loadErr)
			assert.Equal(t, source2.BaseDir, loadErr.Name)

			tests.AssertPgTableExist(ctx, t, conn, "pre_validation_2")
		})
	})
}

func createProjectConfig() project.Configuration {
	return project.Configuration{
		MigrationsTableName: "migrations",
	}
}

func writeUpSQL(t *testing.T, dir string, content string) {
	t.Helper()

	path := filepath.Join(dir, source.UpSQLFilename)
	err := os.WriteFile(path, []byte(content), 0600)

	require.NoError(t, err)
}
