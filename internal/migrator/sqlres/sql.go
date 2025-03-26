package sqlres

import (
	_ "embed"
	"strings"
)

//go:embed ddl.sql
var ddl string

//go:embed register-migration.sql
var registerMigrationQuery string

func DDL(tableName string) string {
	return strings.ReplaceAll(ddl, "_table_name_", tableName)
}

func RegisterMigrationQuery(tableName string) string {
	return strings.ReplaceAll(registerMigrationQuery, "_table_name_", tableName)
}
