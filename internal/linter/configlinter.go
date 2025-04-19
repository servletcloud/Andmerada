package linter

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/servletcloud/Andmerada/internal/schema"
	"github.com/servletcloud/Andmerada/internal/ymlutil"
	"gopkg.in/yaml.v3"
)

type ConfigLinter struct {
	ProjectDir string
}

func (linter *ConfigLinter) Lint(report *Report, relative string, configuration any) bool {
	path := filepath.Join(linter.ProjectDir, relative)

	err := ymlutil.LoadFromFile(path, schema.GetMigrationSchema(), configuration)

	if err == nil {
		return true
	}

	errorMessage := linter.translateError(err)
	report.AddError(errorMessage, relative)

	return false
}

func (linter *ConfigLinter) translateError(err error) string {
	if errors.Is(err, os.ErrNotExist) {
		return "File does not exist"
	} else if schemaError := new(ymlutil.ValidationError); errors.As(err, &schemaError) {
		return fmt.Sprint("Schema validation failed:\n", schemaError)
	} else if yamlError := new(yaml.TypeError); errors.As(err, &yamlError) {
		return yamlError.Error()
	}

	return fmt.Sprint("Failed to read, parse, or validate the migration file:\n", err.Error())
}
