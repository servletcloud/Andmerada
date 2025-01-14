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

		t.Run("No migrations warning", func(t *testing.T) {
			report := runLint(dir, nil)

			assert.Empty(t, report.Errors)
			assertHasError(t, report.Warings, "No migration files found. Create with:")
		})

		_, err := source.Create(dir, "Create users table", timestamp)
		require.NoError(t, err)

		report := runLint(dir, nil)

		assert.Empty(t, report.Errors)
		assert.Empty(t, report.Warings)

		t.Run("No migration.yml file", func(t *testing.T) {
			migrationDir := createTempMigration(t, dir, timestamp2)
			path := filepath.Join(dir, migrationDir, "migration.yml")
			require.NoError(t, os.Remove(path))

			report := runLint(dir, nil)

			assertHasError(t, report.Errors, "File does not exist")
		})

		t.Run("up.sql is missing", func(t *testing.T) {
			migrationDir := createTempMigration(t, dir, timestamp2)
			path := filepath.Join(dir, migrationDir, "up.sql")
			require.NoError(t, os.Remove(path))

			report := runLint(dir, nil)

			assertHasError(t, report.Errors, "File referenced by migration.yml does not exist")
		})

		t.Run("down.sql is missing", func(t *testing.T) {
			migrationDir := createTempMigration(t, dir, timestamp2)
			path := filepath.Join(dir, migrationDir, "down.sql")
			require.NoError(t, os.Remove(path))

			report := runLint(dir, nil)

			assertHasError(t, report.Errors, "File referenced by migration.yml does not exist")
		})

		t.Run("down.sql is missing with down.block=true", func(t *testing.T) {
			migrationDir := createTempMigration(t, dir, timestamp2)
			path := filepath.Join(dir, migrationDir, "down.sql")
			require.NoError(t, os.Remove(path))

			updateConfig(t, filepath.Join(dir, migrationDir, "migration.yml"), func(conf *source.Configuration) {
				conf.Down.Block = true
			})

			report := runLint(dir, nil)

			assert.Empty(t, report.Errors)
			assert.Empty(t, report.Warings)
		})

		t.Run("duplicate migration ID", func(t *testing.T) {
			dupeFolderPath := filepath.Join(dir, "20241225112129_duplicate_migration")
			require.NoError(t, os.Mkdir(dupeFolderPath, osutil.DirPerm0755))

			t.Cleanup(func() {
				require.NoError(t, os.RemoveAll(dupeFolderPath))
			})

			report := runLint(dir, nil)

			assertHasError(t, report.Errors, "Duplicate migration ID")
		})

		t.Run("err of SQL big file size", func(t *testing.T) {
			migrationDir := createTempMigration(t, dir, timestamp2)
			path := filepath.Join(dir, migrationDir, "up.sql")

			stat, err := os.Stat(path)
			require.NoError(t, err)

			lintConfig := &source.LintConfiguration{MaxSQLFileSize: stat.Size() - 1} //nolint:exhaustruct
			report := runLint(dir, lintConfig)

			assertHasError(t, report.Errors, "File is too big:")
		})

		t.Run("warning if there are migrations in the future", func(t *testing.T) {
			_ = createTempMigration(t, dir, timestamp2)

			lintConfig := &source.LintConfiguration{NowUTC: timestamp.Add(-1 * time.Second)} //nolint:exhaustruct
			report := runLint(dir, lintConfig)

			assertHasError(t, report.Warings, "There are migrations with timestamps in the future")
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

func assertHasError(t *testing.T, errors []source.LintError, expectedTitle string) {
	t.Helper()

	found := slices.ContainsFunc(errors, func(lintError source.LintError) bool {
		return strings.Contains(lintError.Title, expectedTitle)
	})

	assert.True(t, found, "No errors contain title: %v\n. Actual errors: %v", expectedTitle, errors)
}

func runLint(dir string, configOverride *source.LintConfiguration) *source.LintReport {
	config := source.LintConfiguration{
		ProjectDir:      dir,
		MaxSQLFileSize:  1 * humanize.KiByte,
		NowUTC:          time.Now(),
		UpSQLTemplate:   "create table users;",
		DownSQLTemplate: "drop table users",
	}

	if configOverride != nil {
		config.MaxSQLFileSize = configOverride.MaxSQLFileSize
		config.NowUTC = configOverride.NowUTC
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
