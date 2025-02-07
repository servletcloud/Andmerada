package descriptions

import (
	_ "embed"
	"strings"
)

const (
	unixNewLine = "\n"
)

//go:embed init.txt
var initRaw []byte

//go:embed lint.txt
var lintRaw []byte

//go:embed create_migration.txt
var crMigrationRaw []byte

//go:embed migrate.txt
var migrateRaw []byte

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

func loadCommandDescription(s []byte) CommandDescription {
	lines := strings.Split(string(s), unixNewLine)

	return CommandDescription{
		Use:   lines[0],
		Short: lines[1],
		Long:  strings.Join(lines[2:], unixNewLine),
	}
}
