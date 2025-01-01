package schema

import (
	_ "embed"
)

//go:embed migration.yml.v1.json
var migrationSchema []byte

//go:embed andmerada.yml.v1.json
var andmeradaSchema []byte

func GetMigrationSchema() string {
	return string(migrationSchema)
}

func GetAndmeradaSchema() string {
	return string(andmeradaSchema)
}
