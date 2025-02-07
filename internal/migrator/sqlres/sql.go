package sqlres

import (
	_ "embed"
	"strings"
)

//go:embed ddl.sql
var ddlBytes []byte

func DDL(appliedMigrationsName string) string {
	ddl := string(ddlBytes)

	return strings.ReplaceAll(ddl, "_applied_migrations_", appliedMigrationsName)
}
