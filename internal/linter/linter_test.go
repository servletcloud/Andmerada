package linter_test

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/servletcloud/Andmerada/internal/linter"
	"github.com/servletcloud/Andmerada/internal/osutil"
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

	t.Run("No migrations warning", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()

		report := runLint(dir, nil)

		assert.Empty(t, report.Errors)
		assertHasError(t, report.Warnings, "No migration files found. Create with:")
	})

	t.Run("A valid migration found", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()

		_, err := source.Create(dir, "Create users table", timestamp)
		require.NoError(t, err)

		report := runLint(dir, nil)

		assert.Empty(t, report.Errors)
		assert.Empty(t, report.Warnings)
	})

	t.Run("No migration.yml file", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()

		migrationDir := createTempMigration(t, dir, timestamp2)
		path := filepath.Join(migrationDir, "migration.yml")
		require.NoError(t, os.Remove(path))

		report := runLint(dir, nil)

		assertHasError(t, report.Errors, "File does not exist")
	})

	t.Run("up.sql is missing", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()

		migrationDir := createTempMigration(t, dir, timestamp2)
		path := filepath.Join(migrationDir, "up.sql")
		require.NoError(t, os.Remove(path))

		report := runLint(dir, nil)

		assertHasError(t, report.Errors, "File referenced by migration.yml does not exist")
	})

	t.Run("down.sql is missing", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()

		migrationDir := createTempMigration(t, dir, timestamp2)
		path := filepath.Join(migrationDir, "down.sql")
		require.NoError(t, os.Remove(path))

		report := runLint(dir, nil)

		assertHasError(t, report.Errors, "File referenced by migration.yml does not exist")
	})

	t.Run("down.sql is missing with down.block=true", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()

		migrationDir := createTempMigration(t, dir, timestamp2)
		path := filepath.Join(migrationDir, "down.sql")
		require.NoError(t, os.Remove(path))

		updateConfig(t, filepath.Join(migrationDir, "migration.yml"), func(conf *source.Configuration) {
			conf.Down.Block = true
		})

		report := runLint(dir, nil)

		assert.Empty(t, report.Errors)
		assert.Empty(t, report.Warnings)
	})

	t.Run("duplicate migration ID", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()

		_, err := source.Create(dir, "Create users table", timestamp)
		require.NoError(t, err)

		dupeFolderPath := filepath.Join(dir, "20241225112129_duplicate_migration")
		require.NoError(t, os.Mkdir(dupeFolderPath, osutil.DirPerm0755))

		report := runLint(dir, nil)

		assertHasError(t, report.Errors, "Duplicate migration ID")
	})

	t.Run("err of SQL big file size", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()

		migrationDir := createTempMigration(t, dir, timestamp2)
		path := filepath.Join(migrationDir, "up.sql")

		stat, err := os.Stat(path)
		require.NoError(t, err)

		lintConfig := &linter.Configuration{MaxSQLFileSize: stat.Size() - 1} //nolint:exhaustruct
		report := runLint(dir, lintConfig)

		assertHasError(t, report.Errors, "File is too big:")
	})

	t.Run("warning if there are migrations in the future", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()

		_ = createTempMigration(t, dir, timestamp2)

		lintConfig := &linter.Configuration{NowUTC: timestamp.Add(-1 * time.Second)} //nolint:exhaustruct
		report := runLint(dir, lintConfig)

		assertHasError(t, report.Warnings, "There are migrations with timestamps in the future")
	})
}

func createTempMigration(t *testing.T, dir string, timestamp time.Time) string {
	t.Helper()

	result, err := source.Create(dir, "Create a bad table", timestamp)
	require.NoError(t, err)

	return result.FullPath
}

func assertHasError(t *testing.T, errors []linter.LintError, expectedTitle string) {
	t.Helper()

	found := slices.ContainsFunc(errors, func(lintError linter.LintError) bool {
		return strings.Contains(lintError.Title, expectedTitle)
	})

	assert.True(t, found, "No errors contain title: %v\n. Actual errors: %v", expectedTitle, errors)
}

func runLint(dir string, configOverride *linter.Configuration) linter.Report {
	config := linter.Configuration{
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

	report := linter.Report{} //nolint:exhaustruct

	if err := linter.Run(config, &report); err != nil {
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

func assertContainsError(t *testing.T, errors []linter.LintError, expectedTitle string) {
	t.Helper()

	found := slices.ContainsFunc(errors, func(lintError linter.LintError) bool {
		return strings.Contains(lintError.Title, expectedTitle)
	})

	assert.True(t, found, "No errors contain title: %v\n. Actual errors: %v", expectedTitle, errors)
}
