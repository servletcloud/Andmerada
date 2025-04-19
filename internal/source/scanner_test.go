package source_test

import (
	"path/filepath"
	"testing"

	"github.com/servletcloud/Andmerada/internal/source"
	"github.com/servletcloud/Andmerada/internal/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScanner_Traverse(t *testing.T) {
	t.Parallel()

	t.Run("list directories and IDs of migrations", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()

		createTestMigrationStructure(t, dir)

		receivedIDs := make([]source.ID, 0)
		receivedNames := make([]string, 0)

		err := source.Traverse(dir, func(id source.ID, name string) bool {
			receivedIDs = append(receivedIDs, id)
			receivedNames = append(receivedNames, name)

			return true
		})
		require.NoError(t, err)

		expectedIDs := []source.ID{
			20230122110947,
			20241225112129,
		}
		assert.Equal(t, expectedIDs, receivedIDs)

		expectedNames := []string{
			"20230122110947_create_users",
			"20241225112129_create_orders",
		}
		assert.Equal(t, expectedNames, receivedNames)
	})

	t.Run("no callbacks invoked on empty dir", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()

		err := source.Traverse(dir, func(_ source.ID, _ string) bool {
			assert.Fail(t, "No IDs must be found in an empty directory")

			return true
		})

		require.NoError(t, err)
	})
}

func TestScanner_ScanAll(t *testing.T) {
	t.Parallel()

	t.Run("list directories and IDs of migrations", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()

		createTestMigrationStructure(t, dir)

		received, err := source.ScanAll(dir)
		require.NoError(t, err)

		expected := map[source.ID]string{
			20230122110947: "20230122110947_create_users",
			20241225112129: "20241225112129_create_orders",
		}
		assert.Equal(t, expected, received)
	})
}

func createTestMigrationStructure(t *testing.T, dir string) {
	t.Helper()

	subDirs := []string{
		"20230122110947_create_users",
		"20241225112129_create_orders",
		".templates",
		".history",
	}

	files := []string{
		"andmerada.yml",
		".gitignore",
		"20230122110947_create_users.txt",
	}

	for _, subDir := range subDirs {
		tests.MkDir(t, filepath.Join(dir, subDir))
	}

	for _, file := range files {
		tests.MkFile(t, filepath.Join(dir, file), "hi")
	}
}
