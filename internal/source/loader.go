package source

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/servletcloud/Andmerada/internal/schema"
	"github.com/servletcloud/Andmerada/internal/ymlutil"
)

type Loader struct {
	MaxSQLFileSize int64
}

func (loader *Loader) LoadSource(dir string, out *Source) error {
	var configuration Configuration

	if err := loader.loadConfiguration(dir, &configuration); err != nil {
		return err
	}

	upSQLPath := filepath.Join(dir, configuration.Up.File)
	upSQL, err := loader.loadSQLFile(upSQLPath)

	if err != nil {
		return err
	}

	downSQL := ""

	if !configuration.Down.Block {
		downSQLPath := filepath.Join(dir, configuration.Down.File)
		downSQL, err = loader.loadSQLFile(downSQLPath)

		if err != nil {
			return err
		}
	}

	out.Configuration = configuration
	out.UpSQL = upSQL
	out.DownSQL = downSQL

	return nil
}

func (loader *Loader) loadConfiguration(dir string, out *Configuration) error {
	path := filepath.Join(dir, MigrationYmlFilename)
	schema := schema.GetMigrationSchema()

	return ymlutil.LoadFromFile(path, schema, out) //nolint:wrapcheck
}

func (loader *Loader) loadSQLFile(path string) (string, error) {
	stat, err := os.Stat(path)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", err //nolint:wrapcheck
		}

		return "", fmt.Errorf("file %q referenced by migration.yml cannot be read: %w", path, err)
	}

	if stat.IsDir() {
		return "", &CannotBeDirError{Name: path}
	}

	if stat.Size() > loader.MaxSQLFileSize {
		return "", &FileTooBigError{Name: path, Size: stat.Size(), Limit: loader.MaxSQLFileSize}
	}

	content, err := os.ReadFile(path)

	return string(content), err
}
