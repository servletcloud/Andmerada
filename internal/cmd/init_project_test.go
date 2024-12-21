package cmd_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/servletcloud/Andmerada/internal/cmd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitProject(t *testing.T) {
	t.Parallel()

	t.Run("Create folder, configuration file, resolves configuration placeholders", func(t *testing.T) {
		t.Parallel()

		targetDir := filepath.Join(t.TempDir(), "migrations/main_db_project")

		require.NoError(t, cmd.InitializeProject(targetDir))

		actualDir, err := os.Stat(targetDir)
		require.NoError(t, err)
		require.True(t, actualDir.IsDir())

		contentBytes, err := os.ReadFile(filepath.Join(targetDir, "andmerada.yml"))
		require.NoError(t, err)

		content := string(contentBytes)
		assert.Contains(t, content, `project: "main_db_project"`)
		assert.NotContains(t, content, "{{")
		assert.NotContains(t, content, "}}")
	})

	t.Run("Returns specific error it target project does already exist", func(t *testing.T) {
		t.Parallel()

		targetDir := t.TempDir()

		require.NoError(t, os.WriteFile(filepath.Join(targetDir, "andmerada.yml"), []byte("hello"), 0600))

		assert.ErrorIs(t, cmd.InitializeProject(targetDir), cmd.ErrConfigFileAlreadyExists)
	})
}
