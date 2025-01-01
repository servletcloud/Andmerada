package project

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"unicode/utf8"

	"github.com/servletcloud/Andmerada/internal/osutil"
	"github.com/servletcloud/Andmerada/internal/resources"
)

const (
	MaxNameLength      = 255
	rootConfigFilename = "andmerada.yml"
)

var (
	ErrConfigFileAlreadyExists = errors.New("configuration file already exists")
	ErrNameExceeds255          = errors.New("name exceeds 255 characters")
)

func Initialize(dir string) error {
	projectName := filepath.Base(dir)
	if utf8.RuneCountInString(projectName) > MaxNameLength {
		return ErrNameExceeds255
	}

	if err := os.MkdirAll(dir, osutil.DirPerm0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	configPath := filepath.Join(dir, rootConfigFilename)

	content := resources.TemplateAndmeradaYml(projectName)

	if err := osutil.WriteFileExcl(configPath, content); err != nil {
		if errors.Is(err, os.ErrExist) {
			return fmt.Errorf("configuration file %s already exists: %w", configPath, ErrConfigFileAlreadyExists)
		}

		return fmt.Errorf("cannot create or write to configuration file %s: %w", configPath, err)
	}

	return nil
}
