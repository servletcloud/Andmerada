package tests

import (
	"os"
	"testing"

	"github.com/servletcloud/Andmerada/internal/osutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func AssertFileContains(t *testing.T, path string, expected string) {
	t.Helper()

	content := ReadFileToString(t, path)

	assert.Contains(t, content, expected)
}

func AssertPlaceholdersResolved(t *testing.T, content string) {
	t.Helper()

	assert.NotContains(t, content, "{{")
	assert.NotContains(t, content, "}}")
}

func MkDir(t *testing.T, path string) {
	t.Helper()

	err := os.Mkdir(path, osutil.DirPerm0755)
	require.NoError(t, err)
}

func ReadFileToString(t *testing.T, path string) string {
	t.Helper()

	bytes, err := os.ReadFile(path)
	require.NoError(t, err)

	return string(bytes)
}
