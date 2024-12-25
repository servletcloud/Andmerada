package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/servletcloud/Andmerada/internal/osutil"
	"github.com/servletcloud/Andmerada/internal/resources"
)

const (
	rootConfigFilename string = "andmerada.yml"
)

var (
	ErrConfigFileAlreadyExists = errors.New("configuration file already exists")
)

func InitializeProject(projectDir string) error {
	if err := os.MkdirAll(projectDir, osutil.DirPerm0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", projectDir, err)
	}

	configPath := filepath.Join(projectDir, rootConfigFilename)

	projectName := filepath.Base(projectDir)
	content := resources.TemplateAndmeradaYml(projectName)

	if err := osutil.WriteFileExcl(configPath, content); err != nil {
		if errors.Is(err, os.ErrExist) {
			return fmt.Errorf("configuration file %s already exists: %w", configPath, ErrConfigFileAlreadyExists)
		}

		return fmt.Errorf("cannot create or write to configuration file %s: %w", configPath, err)
	}

	return nil
}
