package sqlres

import (
	_ "embed"
	"strings"
)

//go:embed ddl.sql
var ddl string

func DDL(appliedMigrationsName string) string {
	return strings.ReplaceAll(ddl, "_applied_migrations_", appliedMigrationsName)
}
