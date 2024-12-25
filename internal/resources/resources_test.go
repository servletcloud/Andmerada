package resources_test

import (
	"testing"

	"github.com/servletcloud/Andmerada/internal/resources"
	"github.com/servletcloud/Andmerada/internal/tests"
	"github.com/stretchr/testify/assert"
)

func TestTemplateAndmeradaYml(t *testing.T) {
	t.Parallel()

	content := resources.TemplateAndmeradaYml("maindb")

	assert.Contains(t, content, `project: "maindb"`)
	tests.AssertPlaceholdersResolved(t, content)
}

func TestTemplateMigrationYml(t *testing.T) {
	t.Parallel()

	content := resources.TemplateMigrationYml("Create users table(!)")

	assert.Contains(t, content, `name: "Create users table(!)"`)
	tests.AssertPlaceholdersResolved(t, content)
}

func TestMsgMigrationCreated(t *testing.T) {
	t.Parallel()

	content := resources.MsgMigrationCreated("20060102150405_create_users")

	assert.Contains(t, content, "Migration successfully created in 20060102150405_create_users")
	tests.AssertPlaceholdersResolved(t, content)
}
