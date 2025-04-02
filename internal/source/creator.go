package source

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
	"unicode/utf8"

	"github.com/servletcloud/Andmerada/internal/osutil"
	"github.com/servletcloud/Andmerada/internal/resources"
)

func create(projectDir string, name string, time time.Time) (CreateSourceResult, error) {
	if utf8.RuneCountInString(name) > MaxNameLength {
		return CreateSourceResult{}, ErrNameExceeds255
	}

	id := NewIDFromTime(time)
	latest, err := verifyIDUnique(id, projectDir)

	if err != nil {
		return CreateSourceResult{}, err
	}

	baseMigrationDir := fmt.Sprintf("%v_%v", id, osutil.NormalizePath(name))
	migrationDir := filepath.Join(projectDir, baseMigrationDir)

	if err = createFiles(name, migrationDir); err != nil {
		return CreateSourceResult{}, err
	}

	return CreateSourceResult{BaseDir: baseMigrationDir, FullPath: migrationDir, Latest: latest}, nil
}

func verifyIDUnique(newID uint64, projectDir string) (bool, error) {
	unique := true
	collidesWith := ""
	latest := true

	err := Traverse(projectDir, func(existingID uint64, name string) bool {
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
			ErrSourceAlreadyExists,
		)

		return latest, err
	}

	return latest, nil
}

func createFiles(name, dir string) error {
	if err := os.Mkdir(dir, osutil.DirPerm0755); err != nil {
		return fmt.Errorf("cannot create a migration directory %v: %w", dir, err)
	}

	configContent := resources.TemplateMigrationYml(name)
	configFilename := filepath.Join(dir, MigrationYmlFilename)

	if err := osutil.WriteFileExcl(configFilename, configContent); err != nil {
		return fmt.Errorf("cannot create %v: %w", configFilename, err)
	}

	upSQLFilename := filepath.Join(dir, UpSQLFilename)
	if err := osutil.WriteFileExcl(upSQLFilename, resources.TemplateUpSQL()); err != nil {
		return fmt.Errorf("cannot create %v: %w", upSQLFilename, err)
	}

	downSQLFilename := filepath.Join(dir, DownSQLFilename)
	if err := osutil.WriteFileExcl(downSQLFilename, resources.TemplateDownSQL()); err != nil {
		return fmt.Errorf("cannot create %v: %w", downSQLFilename, err)
	}

	return nil
}
