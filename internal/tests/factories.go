package tests

import (
	"os"
	"testing"

	"github.com/servletcloud/Andmerada/internal/source"
	"github.com/stretchr/testify/require"
)

func CreateSource(t *testing.T, dir string, title string, timestamp string) source.CreateSourceResult {
	t.Helper()

	id := source.NewIDFromString(timestamp)

	result, err := source.Create(dir, title, id)
	require.NoError(t, err)

	t.Cleanup(func() {
		err := os.RemoveAll(result.FullPath)
		require.NoError(t, err)
	})

	return result
}
