package source

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/servletcloud/Andmerada/internal/schema"
	"github.com/servletcloud/Andmerada/internal/ymlutil"
)

type Loader struct {
	MaxSQLFileSize int64
}

type readFileFunc func(string) ([]byte, error)

func (loader *Loader) LoadSource(dir string, out *Source) error {
	return loader.loadSource(dir, out, os.ReadFile)
}

func (loader *Loader) ValidateSource(dir string, out *Source) error {
	ensureFileRedable := func(path string) error {
		file, err := os.Open(path)

		if err != nil {
			return fmt.Errorf("file %q referenced by migration.yml cannot be read: %w", path, err)
		}

		return file.Close()
	}

	return loader.loadSource(dir, out, func(path string) ([]byte, error) {
		return []byte{}, ensureFileRedable(path)
	})
}

func (loader *Loader) loadSource(dir string, out *Source, readFunc readFileFunc) error {
	if err := loader.loadConfiguration(dir, &out.Configuration); err != nil {
		return err
	}

	return loader.loadSQLFiles(dir, out, readFunc)
}

func (loader *Loader) loadConfiguration(dir string, out *Configuration) error {
	path := filepath.Join(dir, MigrationYmlFilename)
	schema := schema.GetMigrationSchema()

	return ymlutil.LoadFromFile(path, schema, out) //nolint:wrapcheck
}

func (loader *Loader) loadSQLFiles(dir string, out *Source, readFunc readFileFunc) error {
	config := out.Configuration

	upSQL, err := loader.loadSQLFile(dir, config.Up.File, readFunc)

	if err != nil {
		return err
	}

	out.UpSQL = upSQL

	if config.Down.Block {
		return nil
	}

	out.DownSQL, err = loader.loadSQLFile(dir, config.Down.File, readFunc)

	return err
}

func (loader *Loader) loadSQLFile(dir, file string, readFunc readFileFunc) (string, error) {
	path := filepath.Join(dir, file)
	stat, err := os.Stat(path)

	if err != nil {
		return "", err //nolint:wrapcheck
	}

	if stat.IsDir() {
		return "", &CannotBeDirError{Name: path}
	}

	if stat.Size() > loader.MaxSQLFileSize {
		return "", &FileTooBigError{Name: path, Size: stat.Size(), Limit: loader.MaxSQLFileSize}
	}

	content, err := readFunc(path)

	return string(content), err
}
