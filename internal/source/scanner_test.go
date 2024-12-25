package source_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/servletcloud/Andmerada/internal/osutil"
	"github.com/servletcloud/Andmerada/internal/source"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScan(t *testing.T) { //nolint:funlen
	t.Parallel()

	t.Run("list directories and IDs of migrations", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()

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
			err := os.Mkdir(filepath.Join(dir, subDir), osutil.DirPerm0755)
			require.NoError(t, err)
		}

		for _, file := range files {
			err := osutil.WriteFileExcl(filepath.Join(dir, file), "hi")
			require.NoError(t, err)
		}

		receivedIDs := make([]source.MigrationID, 0)
		receivedNames := make([]string, 0)
		err := source.Scan(dir, func(id source.MigrationID, name string) bool {
			receivedIDs = append(receivedIDs, id)
			receivedNames = append(receivedNames, name)

			return true
		})
		require.NoError(t, err)

		expectedIDs := []source.MigrationID{
			source.MigrationID(20060102150405),
			source.MigrationID(20241225112129),
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

		err := source.Scan(dir, func(_ source.MigrationID, _ string) bool {
			assert.Fail(t, "No IDs must be found in an empty directory")

			return true
		})

		require.NoError(t, err)
	})
}
