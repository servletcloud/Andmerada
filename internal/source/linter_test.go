package source_test

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/servletcloud/Andmerada/internal/osutil"
	"github.com/servletcloud/Andmerada/internal/project"
	"github.com/servletcloud/Andmerada/internal/schema"
	"github.com/servletcloud/Andmerada/internal/source"
	"github.com/servletcloud/Andmerada/internal/ymlutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestLint(t *testing.T) { //nolint:funlen
	t.Parallel()

	timestamp, err := time.Parse("20060102150405", "20241225112129")
	require.NoError(t, err)

	timestamp2, err := time.Parse("20060102150405", "20241226122230")
	require.NoError(t, err)

	t.Run("A project with a migration has no errors", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()

		require.NoError(t, project.Initialize(dir))

		_, err := source.Create(dir, "Create users table", timestamp)
		require.NoError(t, err)

		report := runLint(dir, 1*humanize.KiByte)

		assert.Empty(t, report.Errors)
		assert.Empty(t, report.Warings)

		t.Run("No migration.yml file", func(t *testing.T) {
			migrationDir := createTempMigration(t, dir, timestamp2)
			path := filepath.Join(dir, migrationDir, "migration.yml")
			require.NoError(t, os.Remove(path))

			report := runLint(dir, 1*humanize.KiByte)

			assertHasError(t, report, "File does not exist")
		})

		t.Run("migration.yml violates the schema", func(t *testing.T) {
			migrationDir := createTempMigration(t, dir, timestamp2)
			path := filepath.Join(dir, migrationDir, "migration.yml")
			require.NoError(t, os.WriteFile(path, []byte("---"), osutil.FilePerm0644))

			report := runLint(dir, 1*humanize.KiByte)

			assertHasError(t, report, "Schema validation failed")
		})

		t.Run("migration.yml has invalid syntax", func(t *testing.T) {
			migrationDir := createTempMigration(t, dir, timestamp2)
			path := filepath.Join(dir, migrationDir, "migration.yml")
			require.NoError(t, os.WriteFile(path, []byte("bad yaml"), osutil.FilePerm0644))

			report := runLint(dir, 1*humanize.KiByte)

			assertHasError(t, report, "Invalid YAM")
		})

		t.Run("up.sql is missing", func(t *testing.T) {
			migrationDir := createTempMigration(t, dir, timestamp2)
			path := filepath.Join(dir, migrationDir, "up.sql")
			require.NoError(t, os.Remove(path))

			report := runLint(dir, 1*humanize.KiByte)

			assertHasError(t, report, "File referenced by migration.yml does not exist")
		})

		t.Run("down.sql is missing", func(t *testing.T) {
			migrationDir := createTempMigration(t, dir, timestamp2)
			path := filepath.Join(dir, migrationDir, "down.sql")
			require.NoError(t, os.Remove(path))

			report := runLint(dir, 1*humanize.KiByte)

			assertHasError(t, report, "File referenced by migration.yml does not exist")
		})

		t.Run("down.sql is missing with down.block=true", func(t *testing.T) {
			migrationDir := createTempMigration(t, dir, timestamp2)
			path := filepath.Join(dir, migrationDir, "down.sql")
			require.NoError(t, os.Remove(path))

			updateConfig(t, filepath.Join(dir, migrationDir, "migration.yml"), func(conf *source.Configuration) {
				conf.Down.Block = true
			})

			report := runLint(dir, 1*humanize.KiByte)

			assert.Empty(t, report.Errors)
			assert.Empty(t, report.Warings)
		})

		t.Run("duplicate migration ID", func(t *testing.T) {
			dupeFolderPath := filepath.Join(dir, "20241225112129_duplicate_migration")
			require.NoError(t, os.Mkdir(dupeFolderPath, osutil.DirPerm0755))

			t.Cleanup(func() {
				require.NoError(t, os.RemoveAll(dupeFolderPath))
			})

			report := runLint(dir, 1*humanize.KiByte)

			assertHasError(t, report, "Duplicate migration ID")
		})

		t.Run("err of SQL big file size", func(t *testing.T) {
			migrationDir := createTempMigration(t, dir, timestamp2)
			path := filepath.Join(dir, migrationDir, "up.sql")

			stat, err := os.Stat(path)
			require.NoError(t, err)

			report := runLint(dir, stat.Size()-1)

			assertHasError(t, report, "File is too big:")
		})
	})
}

func createTempMigration(t *testing.T, dir string, timestamp time.Time) string {
	t.Helper()

	result, err := source.Create(dir, "Create a bad table", timestamp)
	require.NoError(t, err)

	t.Cleanup(func() {
		err := os.RemoveAll(filepath.Join(dir, result.BaseDir))
		require.NoError(t, err)
	})

	return result.BaseDir
}

func assertHasError(t *testing.T, report *source.LintReport, expectedTitle string) {
	t.Helper()

	found := slices.ContainsFunc(report.Errors, func(lintError source.LintError) bool {
		return strings.Contains(lintError.Title, expectedTitle)
	})

	assert.True(t, found, "No errors contain title: %v", expectedTitle)
}

func runLint(dir string, maxSQLSize int64) *source.LintReport {
	config := source.LintConfiguration{
		ProjectDir:     dir,
		MaxSQLFileSize: maxSQLSize,
	}
	report := new(source.LintReport)

	if err := source.Lint(config, report); err != nil {
		panic(err)
	}

	return report
}

func updateConfig(t *testing.T, path string, callback func(config *source.Configuration)) {
	t.Helper()

	configuration := new(source.Configuration)

	err := ymlutil.LoadFromFile(path, schema.GetMigrationSchema(), configuration)
	require.NoError(t, err)

	callback(configuration)

	toUpdate, err := yaml.Marshal(configuration)
	require.NoError(t, err)

	err = os.WriteFile(path, toUpdate, osutil.FilePerm0644)
	require.NoError(t, err)
}
