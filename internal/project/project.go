package project

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/servletcloud/Andmerada/internal/osutil"
	"github.com/servletcloud/Andmerada/internal/resources"
	"github.com/servletcloud/Andmerada/internal/schema"
	"github.com/servletcloud/Andmerada/internal/ymlutil"
)

const (
	rootConfigFilename = "andmerada.yml"
)

type Project struct {
	Dir           string
	Configuration Configuration
}

type Configuration struct {
	MigrationsTableName string `yaml:"migrations_table_name"`
}

var (
	ErrConfigFileAlreadyExists = errors.New("configuration file already exists")
)

func Initialize(dir string) error {
	if err := os.MkdirAll(dir, osutil.DirPerm0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	configPath := filepath.Join(dir, rootConfigFilename)

	content := resources.TemplateAndmeradaYml()

	if err := osutil.WriteFileExcl(configPath, content); err != nil {
		if errors.Is(err, os.ErrExist) {
			return fmt.Errorf("configuration file %s already exists: %w", configPath, ErrConfigFileAlreadyExists)
		}

		return fmt.Errorf("cannot create or write to configuration file %s: %w", configPath, err)
	}

	return nil
}

func Load(dir string) (Project, error) {
	configFileName := filepath.Join(dir, rootConfigFilename)

	var configuration Configuration

	if err := ymlutil.LoadFromFile(configFileName, schema.GetAndmeradaSchema(), &configuration); err != nil {
		return Project{}, fmt.Errorf("failed to load project configuration file %q: %w", configFileName, err)
	}

	return Project{Dir: dir, Configuration: configuration}, nil
}
