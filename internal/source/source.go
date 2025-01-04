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

type LintError struct {
	Files   []string
	Title   string
	Details string
}

type LintReport struct {
	Errors  []LintError
	Warings []LintError
}

type Configuration struct {
	Name string `yaml:"name"`

	Up struct {
		File string `yaml:"file"`
	} `yaml:"up"`

	Down struct {
		File        string `yaml:"file"`
		Block       bool   `yaml:"block"`
		BlockReason string `yaml:"block_reason"` //nolint:tagliatelle
	} `yaml:"down"`

	Meta map[string]interface{} `yaml:"meta"`
}

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

func Lint(projectDir string, report *LintReport) error {
	return lint(projectDir, report)
}

func Scan(projectDir string, callback func(id MigrationID, name string) bool) error {
	return scan(projectDir, callback)
}

func (report *LintReport) AddError(file, title, details string) {
	report.Errors = append(report.Errors, LintError{
		Title:   title,
		Files:   []string{file},
		Details: details,
	})
}
