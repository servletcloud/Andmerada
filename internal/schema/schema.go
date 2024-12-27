package schema

import (
	_ "embed"
)

//go:embed migration.yml.v1.json
var migrationSchema []byte

func GetMigrationSchema() string {
	return string(migrationSchema)
}
