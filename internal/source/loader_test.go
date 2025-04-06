package source_test

import (
	"testing"

	"github.com/servletcloud/Andmerada/internal/resources"
	"github.com/servletcloud/Andmerada/internal/source"
	"github.com/servletcloud/Andmerada/internal/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadSource(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	loader := source.Loader{MaxSQLFileSize: 1024}
	createdSource := tests.CreateSource(t, dir, "Cr table", "20241225112129")
	sourceDir := createdSource.FullPath

	t.Run("Test loads source", func(t *testing.T) {
		t.Parallel()

		src := source.Source{} //nolint:exhaustruct
		err := loader.LoadSource(sourceDir, &src)
		require.NoError(t, err)

		assert.Equal(t, resources.TemplateUpSQL(), src.UpSQL)
		assert.Equal(t, resources.TemplateDownSQL(), src.DownSQL)

		config := src.Configuration
		assert.Equal(t, "Cr table", config.Name)

		assert.Equal(t, "up.sql", config.Up.File)

		assert.Equal(t, "down.sql", config.Down.File)
		assert.False(t, config.Down.Block)
	})
}
