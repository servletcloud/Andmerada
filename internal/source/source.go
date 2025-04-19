package source

import (
	"errors"
)

type CreateSourceResult struct {
	BaseDir  string
	FullPath string
	Latest   bool
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

func Create(projectDir string, name string, id ID) (CreateSourceResult, error) {
	return create(projectDir, name, id)
}
