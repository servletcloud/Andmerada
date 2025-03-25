package descriptions

import (
	_ "embed"
	"strings"
)

const (
	unixNewLine = "\n"
)

//go:embed init.txt
var initRaw string

//go:embed lint.txt
var lintRaw string

//go:embed create_migration.txt
var crMigrationRaw string

//go:embed migrate.txt
var migrateRaw string

type CommandDescription struct {
	Use   string
	Short string
	Long  string
}

func InitDescription() CommandDescription {
	return loadCommandDescription(initRaw)
}

func CrMigrationDescription() CommandDescription {
	return loadCommandDescription(crMigrationRaw)
}

func LintDescription() CommandDescription {
	return loadCommandDescription(lintRaw)
}

func MigrateDescription() CommandDescription {
	return loadCommandDescription(migrateRaw)
}

func loadCommandDescription(s string) CommandDescription {
	lines := strings.Split(s, unixNewLine)

	return CommandDescription{
		Use:   lines[0],
		Short: lines[1],
		Long:  strings.Join(lines[2:], unixNewLine),
	}
}
