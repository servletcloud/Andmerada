package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/servletcloud/Andmerada/internal/osutil"
	"github.com/servletcloud/Andmerada/internal/resources"
)

const (
	rootConfigFilename string = "andmerada.yml"
)

var (
	errConfigFileAlreadyExists = errors.New("a specific error occurred")
)

func initializeProject(targetDir string) error {
	if err := os.MkdirAll(targetDir, osutil.DirPerm0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", targetDir, err)
	}

	configPath := filepath.Join(targetDir, rootConfigFilename)

	projectName := filepath.Base(targetDir)
	content := strings.ReplaceAll(resources.TemplateAndmeradaYml(), "{{project_name}}", projectName)

	if err := osutil.WriteFile(configPath, content, osutil.O_CREATE_EXCL_WRONLY, osutil.FilePerm0644); err != nil {
		if errors.Is(err, os.ErrExist) {
			return fmt.Errorf("configuration file %s already exists: %w", configPath, errConfigFileAlreadyExists)
		}

		return fmt.Errorf("can not create or write to configuration file %s: %w", configPath, err)
	}

	return nil
}
