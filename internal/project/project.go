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
	MigrationsTableName string `yaml:"migrations_table_name"`
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
