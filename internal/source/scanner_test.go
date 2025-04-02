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

		receivedIDs := make([]uint64, 0)
		receivedNames := make([]string, 0)

		err := source.Traverse(dir, func(id uint64, name string) bool {
			receivedIDs = append(receivedIDs, id)
			receivedNames = append(receivedNames, name)

			return true
		})
		require.NoError(t, err)

		expectedIDs := []uint64{
			20060102150405,
			20241225112129,
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

		err := source.Traverse(dir, func(_ uint64, _ string) bool {
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

		expected := map[uint64]string{
			20060102150405: "20060102150405_create_users",
			20241225112129: "20241225112129_create_orders",
		}
		assert.Equal(t, expected, received)
	})
}

func createTestMigrationStructure(t *testing.T, dir string) {
	t.Helper()

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
		tests.MkDir(t, filepath.Join(dir, subDir))
	}

	for _, file := range files {
		tests.MkFile(t, filepath.Join(dir, file), "hi")
	}
}
