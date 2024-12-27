package cmd

import "errors"

const (
	MaxNameLength = 255

	rootConfigFilename string = "andmerada.yml"
)

var (
	ErrConfigFileAlreadyExists = errors.New("configuration file already exists")
	ErrNameExceeds255          = errors.New("name exceeds 255 characters")
	ErrMigrationAlreadyExists  = errors.New("a migration with the same ID already exists")
)
