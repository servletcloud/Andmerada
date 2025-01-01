package source_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/servletcloud/Andmerada/internal/source"
	"github.com/servletcloud/Andmerada/internal/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateMigration(t *testing.T) {
	t.Parallel()

	name := "Add users table!"
	timestamp, err := time.Parse("20060102150405", "20241225112129")
	require.NoError(t, err)

	t.Run("creates migration file structure", func(t *testing.T) {
		t.Parallel()

		projectDir := t.TempDir()

		result, err := source.Create(projectDir, name, timestamp)
		require.NoError(t, err)

		assert.Equal(t, "20241225112129_add_users_table", result.BaseDir)
		assert.True(t, result.Latest)

		assertMigrationFilesCreated(t, filepath.Join(projectDir, "20241225112129_add_users_table"))
	})

	t.Run("does not create anything when ID collides", func(t *testing.T) {
		t.Parallel()

		projectDir := t.TempDir()

		tests.MkDir(t, filepath.Join(projectDir, "20241225112129_conflicting_migration"))

		_, err = source.Create(projectDir, name, timestamp)
		require.ErrorIs(t, err, source.ErrSourceAlreadyExists)

		migrationDir := filepath.Join(projectDir, "20241225112129_add_users_table")
		assert.NoDirExists(t, migrationDir)
	})

	t.Run("return flag if the new migration is not the latest", func(t *testing.T) {
		t.Parallel()

		projectDir := t.TempDir()

		tests.MkDir(t, filepath.Join(projectDir, "29991225112129_migration_from_year_2999"))

		result, err := source.Create(projectDir, name, timestamp)
		require.NoError(t, err)

		assert.Equal(t, "20241225112129_add_users_table", result.BaseDir)
		assert.False(t, result.Latest)

		assertMigrationFilesCreated(t, filepath.Join(projectDir, "20241225112129_add_users_table"))
	})
}

func assertMigrationFilesCreated(t *testing.T, dir string) {
	t.Helper()

	assert.DirExists(t, dir)

	fileMigrationYml := filepath.Join(dir, "migration.yml")
	fileUpSQL := filepath.Join(dir, "up.sql")
	fileDownSQL := filepath.Join(dir, "down.sql")

	tests.AssertFileContains(t, fileMigrationYml, `name: "Add users table!"`)
	tests.AssertFileContains(t, fileUpSQL, "CREATE TABLE example_table")
	tests.AssertFileContains(t, fileDownSQL, "DROP TABLE example_table")
}
