package source

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/servletcloud/Andmerada/internal/schema"
	"github.com/servletcloud/Andmerada/internal/ymlutil"
	"gopkg.in/yaml.v3"
)

func lint(projectDir string, report *LintReport) error {
	configurations := make([]Configuration, 0)

	err := scan(projectDir, func(_ MigrationID, name string) bool {
		relativePath := filepath.Join(name, MigrationYmlFilename)
		configuration, err := loadConfiguration(filepath.Join(projectDir, relativePath))

		if err == nil {
			configurations = append(configurations, configuration)
		} else {
			report.Errors = append(report.Errors, configurationErrorToLint(relativePath, err))
		}

		return true
	})

	if err != nil {
		return err
	}

	return nil
}

func loadConfiguration(path string) (Configuration, error) {
	var configuration Configuration

	err := ymlutil.LoadFromFile(path, schema.GetMigrationSchema(), &configuration)

	// Intentionally avoid wrapping errors to ensure clean and actionable messages in lint output.
	return configuration, err //nolint:wrapcheck
}

func configurationErrorToLint(file string, err error) LintError {
	title := "Failed to read, parse, or validate the migration file"
	details := err.Error()

	if errors.Is(err, os.ErrNotExist) {
		title = fmt.Sprintf("Migration file does not exist: %q", file)
		details = ""
	} else if errors.Is(err, ymlutil.ErrSchemaValidation) {
		title = "Schema validation failed"
	} else if yamlError := new(yaml.TypeError); errors.As(err, &yamlError) {
		title = "Invalid YAML"
	}

	return LintError{Files: []string{file}, Title: title, Details: details}
}
