package cmd_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/servletcloud/Andmerada/internal/cmd"
	"github.com/servletcloud/Andmerada/internal/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitProject(t *testing.T) {
	t.Parallel()

	t.Run("Create directory, configuration file, resolves configuration placeholders", func(t *testing.T) {
		t.Parallel()

		projectDir := filepath.Join(t.TempDir(), "migrations/main_db_project")

		require.NoError(t, cmd.InitializeProject(projectDir))

		assert.DirExists(t, projectDir)

		tests.AssertFileContains(t, filepath.Join(projectDir, "andmerada.yml"), `project: "main_db_project"`)
	})

	t.Run("Returns specific error it target project does already exist", func(t *testing.T) {
		t.Parallel()

		projectDir := t.TempDir()

		require.NoError(t, os.WriteFile(filepath.Join(projectDir, "andmerada.yml"), []byte("hello"), 0600))

		assert.ErrorIs(t, cmd.InitializeProject(projectDir), cmd.ErrConfigFileAlreadyExists)
	})
}
