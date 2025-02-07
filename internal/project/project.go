package project

import "errors"

const (
	MaxNameLength      = 255
	rootConfigFilename = "andmerada.yml"
)

type Project struct {
	Dir           string
	Configuration Configuration
}

type Configuration struct {
	Name string `yaml:"name"`

	TableNames struct {
		AppliedMigrations string `yaml:"applied_migrations"` //nolint:tagliatelle
	} `yaml:"table_names"` //nolint:tagliatelle
}

var (
	ErrConfigFileAlreadyExists = errors.New("configuration file already exists")
	ErrNameExceeds255          = errors.New("name exceeds 255 characters")
)

func Initialize(dir string) error {
	return initialize(dir)
}

func Load(dir string) (Project, error) {
	return load(dir)
}
