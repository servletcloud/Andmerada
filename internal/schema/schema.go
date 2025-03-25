package schema

import (
	_ "embed"
)

//go:embed migration.yml.v1.json
var migrationSchema string

//go:embed andmerada.yml.v1.json
var andmeradaSchema string

func GetMigrationSchema() string {
	return migrationSchema
}

func GetAndmeradaSchema() string {
	return andmeradaSchema
}
