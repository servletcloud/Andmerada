package tests

import (
	"os"
	"testing"

	"github.com/servletcloud/Andmerada/internal/osutil"
	"github.com/stretchr/testify/require"
)

func MkFile(t *testing.T, path string, content string) {
	t.Helper()

	t.Cleanup(func() {
		err := os.Remove(path)
		require.NoError(t, err)
	})

	err := osutil.WriteFileExcl(path, content)
	require.NoError(t, err)
}

func MkDir(t *testing.T, path string) {
	t.Helper()

	t.Cleanup(func() {
		err := os.RemoveAll(path)
		require.NoError(t, err)
	})

	err := os.Mkdir(path, osutil.DirPerm0755)
	require.NoError(t, err)
}

func ReadFileToString(t *testing.T, path string) string {
	t.Helper()

	bytes, err := os.ReadFile(path)
	require.NoError(t, err)

	return string(bytes)
}
