package source

import (
	"errors"
	"time"
)

type CreateSourceResult struct {
	BaseDir string
	Latest  bool
}

type MigrationID uint64

const (
	MaxNameLength = 255

	MigrationYmlFilename = "migration.yml"
	UpSQLFilename        = "up.sql"
	DownSQLFilename      = "down.sql"

	EmptyMigrationID = MigrationID(0)
)

var (
	ErrNameExceeds255      = errors.New("name exceeds 255 characters")
	ErrSourceAlreadyExists = errors.New("a migration with the same ID already exists")
)

func NewIDFromTime(t time.Time) MigrationID {
	return newIDFromTime(t)
}

func NewIDFromString(str string) MigrationID {
	return newIDFromString(str)
}

func Create(projectDir string, name string, time time.Time) (CreateSourceResult, error) {
	return create(projectDir, name, time)
}

func Scan(projectDir string, callback func(id MigrationID, name string) bool) error {
	return scan(projectDir, callback)
}
