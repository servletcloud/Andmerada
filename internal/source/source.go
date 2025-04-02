package source

import (
	"errors"
	"time"
)

type CreateSourceResult struct {
	BaseDir  string
	FullPath string
	Latest   bool
}

type LintError struct {
	Title string
	Files []string
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
		BlockReason string `yaml:"block_reason"`
	} `yaml:"down"`

	Meta map[string]any `yaml:"meta"`
}

type Source struct {
	Configuration Configuration
	UpSQL         string
	DownSQL       string
}

type LintConfiguration struct {
	ProjectDir      string
	MaxSQLFileSize  int64
	NowUTC          time.Time
	UpSQLTemplate   string
	DownSQLTemplate string
}

const (
	MaxNameLength = 255

	MigrationYmlFilename = "migration.yml"
	UpSQLFilename        = "up.sql"
	DownSQLFilename      = "down.sql"

	EmptyMigrationID = uint64(0)
)

var (
	ErrNameExceeds255      = errors.New("name exceeds 255 characters")
	ErrSourceAlreadyExists = errors.New("a migration with the same ID already exists")
)

func NewIDFromTime(t time.Time) uint64 {
	return newIDFromTime(t)
}

func NewIDFromString(str string) uint64 {
	return newIDFromString(str)
}

func Create(projectDir string, name string, time time.Time) (CreateSourceResult, error) {
	return create(projectDir, name, time)
}

func Lint(conf LintConfiguration, report *LintReport) error {
	linter := &linter{
		LintConfiguration: conf,
		report:            report,
	}

	return linter.lint()
}
