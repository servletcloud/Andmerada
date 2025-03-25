package sqlres

import (
	_ "embed"
	"strings"
)

//go:embed ddl.sql
var ddl string

//go:embed register-migration.sql
var registerMigrationQuery string

func DDL(appliedMigrationsName string) string {
	return strings.ReplaceAll(ddl, "_applied_migrations_", appliedMigrationsName)
}

func RegisterMigrationQuery(tableName string) string {
	return strings.ReplaceAll(registerMigrationQuery, "__table_name__", tableName)
}
