package source_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/servletcloud/Andmerada/internal/osutil"
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

func TestNewIDFromTime(t *testing.T) {
	t.Parallel()

	timestamp, err := time.Parse("20060102150405", "20241225112129")
	require.NoError(t, err)

	assert.Equal(t, source.MigrationID(20241225112129), source.NewIDFromTime(timestamp))
}

func TestNewIDFromString(t *testing.T) {
	t.Parallel()

	assert.Equal(t, source.MigrationID(20060102150405), source.NewIDFromString("20060102150405_create_users"))
	assert.Equal(t, source.EmptyMigrationID, source.NewIDFromString("2006010215040_create_users"))
	assert.Equal(t, source.EmptyMigrationID, source.NewIDFromString("200601021504056_create_users"))
	assert.Equal(t, source.EmptyMigrationID, source.NewIDFromString("20060102150405create_users"))
}

func TestScan(t *testing.T) { //nolint:funlen
	t.Parallel()

	t.Run("list directories and IDs of migrations", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()

		subDirs := []string{
			"20060102150405_create_users",
			"20241225112129_create_orders",
			".templates",
			".history",
		}

		files := []string{
			"andmerada.yml",
			".gitignore",
			"20060102150405_create_users.txt",
		}

		for _, subDir := range subDirs {
			err := os.Mkdir(filepath.Join(dir, subDir), osutil.DirPerm0755)
			require.NoError(t, err)
		}

		for _, file := range files {
			err := osutil.WriteFileExcl(filepath.Join(dir, file), "hi")
			require.NoError(t, err)
		}

		receivedIDs := make([]source.MigrationID, 0)
		receivedNames := make([]string, 0)
		err := source.Scan(dir, func(id source.MigrationID, name string) bool {
			receivedIDs = append(receivedIDs, id)
			receivedNames = append(receivedNames, name)

			return true
		})
		require.NoError(t, err)

		expectedIDs := []source.MigrationID{
			source.MigrationID(20060102150405),
			source.MigrationID(20241225112129),
		}
		assert.Equal(t, expectedIDs, receivedIDs)

		expectedNames := []string{
			"20060102150405_create_users",
			"20241225112129_create_orders",
		}
		assert.Equal(t, expectedNames, receivedNames)
	})

	t.Run("no callbacks invoked on empty dir", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()

		err := source.Scan(dir, func(_ source.MigrationID, _ string) bool {
			assert.Fail(t, "No IDs must be found in an empty directory")

			return true
		})

		require.NoError(t, err)
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
