package tests

import (
	"os"
	"testing"
	"time"

	"github.com/servletcloud/Andmerada/internal/source"
	"github.com/stretchr/testify/require"
)

func CreateSource(t *testing.T, dir string, title string, timestamp string) source.CreateSourceResult {
	t.Helper()

	timeParsed, err := time.Parse("20060102150405", timestamp)
	require.NoError(t, err)

	result, err := source.Create(dir, title, timeParsed)
	require.NoError(t, err)

	t.Cleanup(func() {
		err := os.RemoveAll(result.FullPath)
		require.NoError(t, err)
	})

	return result
}
