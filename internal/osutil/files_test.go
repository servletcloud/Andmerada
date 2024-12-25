package osutil_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/servletcloud/Andmerada/internal/osutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteFileExcl(t *testing.T) {
	t.Parallel()

	t.Run("Creates a new file with content", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(t.TempDir(), "testfile.txt")

		require.NoError(t, osutil.WriteFileExcl(path, "Hello, world"))

		actual, err := os.ReadFile(path)
		require.NoError(t, err)

		assert.Equal(t, "Hello, world", string(actual))
	})

	t.Run("Fails if O_EXCL and file exists", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(t.TempDir(), "testfile.txt")

		require.NoError(t, osutil.WriteFileExcl(path, "Hello, world"))
		assert.ErrorIs(
			t,
			osutil.WriteFileExcl(path, "Hello, world"),
			os.ErrExist,
		)
	})
}

func TestGetwdOrPanic(t *testing.T) {
	t.Parallel()

	currentDir := osutil.GetwdOrPanic()

	assert.NotNil(t, currentDir)
}

func TestNormalizePath(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "add_users_table", osutil.NormalizePath("Add users table"))
	assert.Equal(t, "add_profiles_table", osutil.NormalizePath("  Add profiles table   "))
	assert.Equal(t, "add_flags_table", osutil.NormalizePath("Add   FLAGS  Table"))
	assert.Equal(t, "add_type_column", osutil.NormalizePath("~ADD!!Type^^Column@@"))
	assert.Equal(t, "add.one-two_three", osutil.NormalizePath("add.one-two_three"))
	assert.Equal(t, "", osutil.NormalizePath("~!@#$%^&*()"))
}
