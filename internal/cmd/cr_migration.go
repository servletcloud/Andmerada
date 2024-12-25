package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/servletcloud/Andmerada/internal/osutil"
	"github.com/servletcloud/Andmerada/internal/resources"
	"github.com/servletcloud/Andmerada/internal/source"
)

const NameMaxLength = 1000

var ErrMigrationAlreadyExists = errors.New("a migration with the same ID already exists")

type CreateMigrationResult struct {
	BaseDir string
	Latest  bool
}

func CreateMigration(projectDir string, name string, time time.Time) (CreateMigrationResult, error) {
	id := source.NewIDFromTime(time) //nolint:varnamelen
	latest, err := verifyNewID(id, projectDir)

	if err != nil {
		return CreateMigrationResult{}, err
	}

	baseMigrationDir := fmt.Sprintf("%v_%v", id, osutil.NormalizePath(name))
	migrationDir := filepath.Join(projectDir, baseMigrationDir)

	if err = createMigration(name, migrationDir); err != nil {
		return CreateMigrationResult{}, err
	}

	return CreateMigrationResult{BaseDir: baseMigrationDir, Latest: latest}, nil
}

func verifyNewID(newID source.MigrationID, projectDir string) (bool, error) {
	unique := true
	collidesWith := ""
	latest := true

	err := source.Scan(projectDir, func(existingID source.MigrationID, name string) bool {
		if newID == existingID {
			unique = false
			collidesWith = name
		}

		if existingID > newID {
			latest = false
		}

		return unique
	})

	if err != nil {
		return latest, fmt.Errorf("failed to scan %s directory for existing migrations: %w", projectDir, err)
	}

	if !unique {
		err = fmt.Errorf(
			"an existing migration %v with the same ID %v already exists: %w",
			collidesWith,
			newID,
			ErrMigrationAlreadyExists,
		)

		return latest, err
	}

	return latest, nil
}

func createMigration(name, dir string) error {
	if err := os.Mkdir(dir, osutil.DirPerm0755); err != nil {
		return fmt.Errorf("cannot create a migration directory %v: %w", dir, err)
	}

	configContent := resources.TemplateMigrationYml(name)
	configFilename := filepath.Join(dir, source.MigrationYmlFilename)

	if err := osutil.WriteFileExcl(configFilename, configContent); err != nil {
		return fmt.Errorf("cannot create %v: %w", configFilename, err)
	}

	upSQLFilename := filepath.Join(dir, source.UpSQLFilename)
	if err := osutil.WriteFileExcl(upSQLFilename, resources.TemplateUpSQL()); err != nil {
		return fmt.Errorf("cannot create %v: %w", upSQLFilename, err)
	}

	downSQLFilename := filepath.Join(dir, source.DownSQLFilename)
	if err := osutil.WriteFileExcl(downSQLFilename, resources.TemplateDownSQL()); err != nil {
		return fmt.Errorf("cannot create %v: %w", downSQLFilename, err)
	}

	return nil
}
