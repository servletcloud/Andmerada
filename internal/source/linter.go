package source

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/servletcloud/Andmerada/internal/schema"
	"github.com/servletcloud/Andmerada/internal/ymlutil"
	"gopkg.in/yaml.v3"
)

func lint(projectDir string, report *LintReport) error {
	configurations := make([]Configuration, 0)

	err := scan(projectDir, func(_ MigrationID, name string) bool {
		migrationYmlPath := filepath.Join(name, MigrationYmlFilename)
		configuration, ok := loadConfiguration(projectDir, migrationYmlPath, report)

		if !ok {
			return true
		}

		upSQLExists := resolveSQLFile(projectDir, filepath.Join(name, configuration.Up.File), report)

		if !configuration.Down.Block {
			resolveSQLFile(projectDir, filepath.Join(name, configuration.Down.File), report)
		}

		if !upSQLExists {
			return true
		}

		configurations = append(configurations, configuration)

		return true
	})

	if err != nil {
		return err
	}

	return nil
}

func loadConfiguration(dir string, relative string, report *LintReport) (Configuration, bool) {
	var configuration Configuration

	path := filepath.Join(dir, relative)

	if err := ymlutil.LoadFromFile(path, schema.GetMigrationSchema(), &configuration); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			report.AddError(relative, "File does not exist", "")
		} else if errors.Is(err, ymlutil.ErrSchemaValidation) {
			report.AddError(relative, "Schema validation failed", err.Error())
		} else if yamlError := new(yaml.TypeError); errors.As(err, &yamlError) {
			report.AddError(relative, "Invalid YAML", err.Error())
		} else {
			report.AddError(relative, "Failed to read, parse, or validate the migration file", err.Error())
		}

		return configuration, false
	}

	return configuration, true
}

func resolveSQLFile(dir string, relative string, report *LintReport) bool {
	stat, err := os.Stat(filepath.Join(dir, relative))

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			report.AddError(relative, "File referenced by migration.yml does not exist", "")
		} else {
			report.AddError(relative, "File referenced by migration.yml cannot be read", err.Error())
		}

		return false
	}

	if stat.IsDir() {
		report.AddError(relative, "Must be a file, but is a directory", "")

		return false
	}

	return true
}
