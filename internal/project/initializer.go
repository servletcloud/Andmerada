package project

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/servletcloud/Andmerada/internal/osutil"
	"github.com/servletcloud/Andmerada/internal/resources"
)

func initialize(dir string) error {
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
