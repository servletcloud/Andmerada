package osutil_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/servletcloud/Andmerada/internal/osutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteFile(t *testing.T) {
	t.Parallel()

	t.Run("Creates a new file with content", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(t.TempDir(), "testfile.txt")

		require.NoError(t, osutil.WriteFile(path, "Hello, world", osutil.O_CREATE_EXCL_WRONLY, osutil.FilePerm0644))

		actual, err := os.ReadFile(path)
		require.NoError(t, err)

		assert.Equal(t, "Hello, world", string(actual))
	})

	t.Run("Fails if O_EXCL and file exists", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(t.TempDir(), "testfile.txt")

		require.NoError(t, osutil.WriteFile(path, "Hello, world", osutil.O_CREATE_EXCL_WRONLY, osutil.FilePerm0644))
		assert.ErrorIs(
			t,
			osutil.WriteFile(path, "Hello, world", osutil.O_CREATE_EXCL_WRONLY, osutil.FilePerm0644),
			os.ErrExist,
		)
	})
}
