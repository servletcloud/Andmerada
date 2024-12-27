package resources_test

import (
	"encoding/json"
	"testing"

	"github.com/servletcloud/Andmerada/internal/resources"
	"github.com/servletcloud/Andmerada/internal/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
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

func TestMigrationYMLTemplateMatchesSchema(t *testing.T) {
	t.Parallel()

	yamlFile := resources.TemplateMigrationYml("Create users table")

	var yamlData map[string]interface{}

	require.NoError(t, yaml.Unmarshal([]byte(yamlFile), &yamlData))

	jsonData, err := json.Marshal(yamlData)
	require.NoError(t, err)

	schemaLoader := gojsonschema.NewReferenceLoader("file://./../../api/schema/migration.yml.v1.json")
	documentLoader := gojsonschema.NewBytesLoader(jsonData)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	require.NoError(t, err)

	assert.True(t, result.Valid(), "migration.yml template does not match the schema")

	for _, desc := range result.Errors() {
		t.Errorf("Schema validation error: %s", desc)
	}
}

func TestMsgMigrationCreated(t *testing.T) {
	t.Parallel()

	content := resources.MsgMigrationCreated("20060102150405_create_users")

	assert.Contains(t, content, "Migration successfully created at 20060102150405_create_users")
	tests.AssertPlaceholdersResolved(t, content)
}
