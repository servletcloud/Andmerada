package source

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
	"unicode/utf8"

	"github.com/servletcloud/Andmerada/internal/osutil"
	"github.com/servletcloud/Andmerada/internal/resources"
)

type CreateSourceResult struct {
	BaseDir string
	Latest  bool
}

const (
	MaxNameLength = 255

	MigrationYmlFilename = "migration.yml"
	UpSQLFilename        = "up.sql"
	DownSQLFilename      = "down.sql"
)

var (
	ErrNameExceeds255      = errors.New("name exceeds 255 characters")
	ErrSourceAlreadyExists = errors.New("a migration with the same ID already exists")
)

func Create(projectDir string, name string, time time.Time) (CreateSourceResult, error) {
	if utf8.RuneCountInString(name) > MaxNameLength {
		return CreateSourceResult{}, ErrNameExceeds255
	}

	id := NewIDFromTime(time) //nolint:varnamelen
	latest, err := verifyNewID(id, projectDir)

	if err != nil {
		return CreateSourceResult{}, err
	}

	baseMigrationDir := fmt.Sprintf("%v_%v", id, osutil.NormalizePath(name))
	migrationDir := filepath.Join(projectDir, baseMigrationDir)

	if err = createMigration(name, migrationDir); err != nil {
		return CreateSourceResult{}, err
	}

	return CreateSourceResult{BaseDir: baseMigrationDir, Latest: latest}, nil
}

func verifyNewID(newID MigrationID, projectDir string) (bool, error) {
	unique := true
	collidesWith := ""
	latest := true

	err := Scan(projectDir, func(existingID MigrationID, name string) bool {
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

func createMigration(name, dir string) error {
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
