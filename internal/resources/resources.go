package resources

import (
	_ "embed"
	"strings"
)

//go:embed template.andmerada.yml
var templateAndmeradaYml []byte

//go:embed template.migration.yml
var templateMigrationYml []byte

//go:embed template.up.sql
var templateUpSQL []byte

//go:embed template.down.sql
var templateDownSQL []byte

//go:embed msg_init_completed.txt
var msgInitCompleted []byte

//go:embed msg_err_project_exists.txt
var msgErrProjectExists []byte

//go:embed msg_migration_created.txt
var msgMigrationCreated []byte

//go:embed msg_migration_not_latest.txt
var msgMigrationNotLatest []byte

func TemplateAndmeradaYml(projectName string) string {
	return strings.ReplaceAll(string(templateAndmeradaYml), "{{project_name}}", projectName)
}

func TemplateMigrationYml(name string) string {
	return strings.ReplaceAll(string(templateMigrationYml), "{{name}}", name)
}

func TemplateUpSQL() string {
	return string(templateUpSQL)
}

func TemplateDownSQL() string {
	return string(templateDownSQL)
}

func MsgInitCompleted() string {
	return string(msgInitCompleted)
}

func MsgErrProjectExists() string {
	return string(msgErrProjectExists)
}

func MsgMigrationCreated(dir string) string {
	return strings.ReplaceAll(string(msgMigrationCreated), "{{dir}}", dir)
}

func MsgMigrationNotLatest() string {
	return string(msgMigrationNotLatest)
}
