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
	idToName := make(map[MigrationID][]string)

	err := scan(projectDir, func(id MigrationID, name string) bool {
		idToName[id] = append(idToName[id], name)

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

	detectDuplicateIDs(idToName, report)

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
		report.AddError(relative, "Must be a file but is a directory", "")

		return false
	}

	return true
}

func detectDuplicateIDs(idToNames map[MigrationID][]string, report *LintReport) {
	for id, names := range idToNames { //nolint:varnamelen
		if len(names) <= 1 {
			continue
		}

		err := LintError{
			Files:   names,
			Title:   fmt.Sprintf("Duplicate migration ID: %v", id),
			Details: "Ensure each migration has a unique timestamp-based ID.",
		}

		report.Errors = append(report.Errors, err)
	}
}
