package resources

import (
	_ "embed"
	"strings"
)

//go:embed template.andmerada.yml
var templateAndmeradaYml string

//go:embed template.migration.yml
var templateMigrationYml string

//go:embed template.up.sql
var templateUpSQL string

//go:embed template.down.sql
var templateDownSQL string

//go:embed msg_init_completed.txt
var msgInitCompleted string

//go:embed msg_err_project_exists.txt
var msgErrProjectExists string

//go:embed msg_migration_created.txt
var msgMigrationCreated string

//go:embed msg_migration_not_latest.txt
var msgMigrationNotLatest string

func TemplateAndmeradaYml(projectName string) string {
	return strings.ReplaceAll(templateAndmeradaYml, "{{project_name}}", projectName)
}

func TemplateMigrationYml(name string) string {
	return strings.ReplaceAll(templateMigrationYml, "{{name}}", name)
}

func TemplateUpSQL() string {
	return templateUpSQL
}

func TemplateDownSQL() string {
	return templateDownSQL
}

func MsgInitCompleted() string {
	return msgInitCompleted
}

func MsgErrProjectExists() string {
	return msgErrProjectExists
}

func MsgMigrationCreated(dir string) string {
	return strings.ReplaceAll(msgMigrationCreated, "{{dir}}", dir)
}

func MsgMigrationNotLatest() string {
	return msgMigrationNotLatest
}
