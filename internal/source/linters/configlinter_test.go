package linters_test

import (
	"path/filepath"
	"testing"

	"github.com/servletcloud/Andmerada/internal/osutil"
	"github.com/servletcloud/Andmerada/internal/resources"
	"github.com/servletcloud/Andmerada/internal/source"
	"github.com/servletcloud/Andmerada/internal/source/linters"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigLinter(t *testing.T) { //nolint:funlen
	t.Parallel()

	t.Run("Config exists and is valid", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := filepath.Join(dir, "migration.yml")

		err := osutil.WriteFileExcl(path, resources.TemplateMigrationYml("create users table"))
		require.NoError(t, err)

		report := new(TestLintReport)
		linter := &linters.ConfigLinter{Reporter: report, ProjectDir: dir}

		configuration := new(source.Configuration)
		valid := linter.Lint("migration.yml", configuration)

		require.True(t, valid)
		assert.Empty(t, report.errors)
		assert.Empty(t, report.warnings)
		assert.Equal(t, "create users table", configuration.Name)
	})

	t.Run("No config file", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()

		report := new(TestLintReport)
		linter := &linters.ConfigLinter{Reporter: report, ProjectDir: dir}

		configuration := new(source.Configuration)
		valid := linter.Lint("migration.yml", configuration)

		require.False(t, valid)
		assert.Contains(t, report.errors, "File does not exist")
		assert.Empty(t, report.warnings)
	})

	t.Run("YAML schema is invalid", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := filepath.Join(dir, "migration.yml")

		err := osutil.WriteFileExcl(path, resources.TemplateMigrationYml(""))
		require.NoError(t, err)

		report := new(TestLintReport)
		linter := &linters.ConfigLinter{Reporter: report, ProjectDir: dir}

		configuration := new(source.Configuration)
		valid := linter.Lint("migration.yml", configuration)

		expectedError := "Schema validation failed:\n- name: String length must be greater than or equal to 1"

		require.False(t, valid)
		assert.Contains(t, report.errors, expectedError)
		assert.Empty(t, report.warnings)
	})

	t.Run("YAML has bad syntax", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := filepath.Join(dir, "migration.yml")

		err := osutil.WriteFileExcl(path, "Not a yaml")
		require.NoError(t, err)

		report := new(TestLintReport)
		linter := &linters.ConfigLinter{Reporter: report, ProjectDir: dir}

		configuration := new(source.Configuration)
		valid := linter.Lint("migration.yml", configuration)

		expectedError := "yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `Not a yaml` into source.Configuration"

		require.False(t, valid)
		assert.Contains(t, report.errors, expectedError)
		assert.Empty(t, report.warnings)
	})
}
