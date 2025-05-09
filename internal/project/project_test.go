package project_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/servletcloud/Andmerada/internal/osutil"
	"github.com/servletcloud/Andmerada/internal/project"
	"github.com/servletcloud/Andmerada/internal/ymlutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestInitialize(t *testing.T) {
	t.Parallel()

	t.Run("Create directory, configuration file", func(t *testing.T) {
		t.Parallel()

		projectDir := filepath.Join(t.TempDir(), "migrations/main_db_project")

		require.NoError(t, project.Initialize(projectDir))

		assert.DirExists(t, projectDir)
		assert.FileExists(t, filepath.Join(projectDir, "andmerada.yml"))
	})

	t.Run("Returns specific error it target project does already exist", func(t *testing.T) {
		t.Parallel()

		projectDir := t.TempDir()

		require.NoError(t, os.WriteFile(filepath.Join(projectDir, "andmerada.yml"), []byte("hello"), 0600))

		assert.ErrorIs(t, project.Initialize(projectDir), project.ErrConfigFileAlreadyExists)
	})
}

//nolint:funlen
func TestLoad(t *testing.T) {
	t.Parallel()

	t.Run("create and load the project", func(t *testing.T) {
		t.Parallel()

		projectDir := t.TempDir()

		require.NoError(t, project.Initialize(projectDir))

		project, err := project.Load(projectDir)
		require.NoError(t, err)

		configuration := project.Configuration

		assert.Equal(t, projectDir, project.Dir)
		assert.Equal(t, "migrations", configuration.MigrationsTableName)
	})

	t.Run("returns os.ErrNotExist when project dir does not exist", func(t *testing.T) {
		t.Parallel()

		projectDir := t.TempDir()

		require.NoError(t, os.RemoveAll(projectDir))

		_, err := project.Load(projectDir)
		assert.ErrorIs(t, err, os.ErrNotExist)
	})

	t.Run("returns os.ErrNotExist when configuration does not exist", func(t *testing.T) {
		t.Parallel()

		projectDir := t.TempDir()

		_, err := project.Load(projectDir)
		assert.ErrorIs(t, err, os.ErrNotExist)
	})

	t.Run("returns os.ErrNotExist when configuration cannot be parsed", func(t *testing.T) {
		t.Parallel()

		projectDir := t.TempDir()
		configPath := filepath.Join(projectDir, "andmerada.yml")

		require.NoError(t, osutil.WriteFileExcl(configPath, "bad yml"))

		_, err := project.Load(projectDir)

		var yamlError *yaml.TypeError

		assert.ErrorAs(t, err, &yamlError)
	})

	t.Run("returns ymlutil.ErrSchemaValidation when JSON schema is invalid", func(t *testing.T) {
		t.Parallel()

		projectDir := t.TempDir()
		configPath := filepath.Join(projectDir, "andmerada.yml")

		require.NoError(t, osutil.WriteFileExcl(configPath, "---"))

		_, err := project.Load(projectDir)

		var validationError *ymlutil.ValidationError

		assert.ErrorAs(t, err, &validationError)
	})
}
