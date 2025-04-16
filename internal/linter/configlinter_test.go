package linter_test

import (
	"path/filepath"
	"testing"

	"github.com/servletcloud/Andmerada/internal/linter"
	"github.com/servletcloud/Andmerada/internal/osutil"
	"github.com/servletcloud/Andmerada/internal/resources"
	"github.com/servletcloud/Andmerada/internal/source"
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

		report := linter.Report{} //nolint:exhaustruct
		linter := &linter.ConfigLinter{ProjectDir: dir}

		configuration := new(source.Configuration)
		valid := linter.Lint(&report, "migration.yml", configuration)

		require.True(t, valid)
		assert.Empty(t, report.Errors)
		assert.Empty(t, report.Warnings)
		assert.Equal(t, "create users table", configuration.Name)
	})

	t.Run("No config file", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()

		report := linter.Report{} //nolint:exhaustruct
		linter := &linter.ConfigLinter{ProjectDir: dir}

		configuration := new(source.Configuration)
		valid := linter.Lint(&report, "migration.yml", configuration)

		require.False(t, valid)
		assertContainsError(t, report.Errors, "File does not exist")
		assert.Empty(t, report.Warnings)
	})

	t.Run("YAML schema is invalid", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := filepath.Join(dir, "migration.yml")

		err := osutil.WriteFileExcl(path, resources.TemplateMigrationYml(""))
		require.NoError(t, err)

		report := linter.Report{} //nolint:exhaustruct
		linter := &linter.ConfigLinter{ProjectDir: dir}

		configuration := new(source.Configuration)
		valid := linter.Lint(&report, "migration.yml", configuration)

		expectedError := "Schema validation failed:\n- name: String length must be greater than or equal to 1"

		require.False(t, valid)
		assertContainsError(t, report.Errors, expectedError)
		assert.Empty(t, report.Warnings)
	})

	t.Run("YAML has bad syntax", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := filepath.Join(dir, "migration.yml")

		err := osutil.WriteFileExcl(path, "Not a yaml")
		require.NoError(t, err)

		report := linter.Report{} //nolint:exhaustruct
		linter := &linter.ConfigLinter{ProjectDir: dir}

		configuration := new(source.Configuration)
		valid := linter.Lint(&report, "migration.yml", configuration)

		expectedError := "yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `Not a yaml` into source.Configuration"

		require.False(t, valid)
		assertContainsError(t, report.Errors, expectedError)
		assert.Empty(t, report.Warnings)
	})
}
