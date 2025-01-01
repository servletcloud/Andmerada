package ymlutil

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
)

func validate(yml []byte, schema string) (*gojsonschema.Result, error) {
	json, err := ymlToJSON(yml)
	if err != nil {
		return nil, err
	}

	schemaLoader := gojsonschema.NewStringLoader(schema)
	documentLoader := gojsonschema.NewBytesLoader(json)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)

	if err != nil {
		return nil, fmt.Errorf("cannot validate JSON schema in YML: %w", err)
	}

	return result, nil
}

func fmtValidationErrors(result *gojsonschema.Result) string {
	if result.Valid() {
		return "No validation errors."
	}

	var stringBuilder strings.Builder

	stringBuilder.WriteString("Validation failed with the following errors:\n")

	for _, desc := range result.Errors() {
		stringBuilder.WriteString(fmt.Sprintf("- %s\n", desc.String()))
	}

	return stringBuilder.String()
}

func ymlToJSON(yml []byte) ([]byte, error) {
	var yamlData map[string]interface{}

	if err := yaml.Unmarshal(yml, &yamlData); err != nil {
		return nil, fmt.Errorf("cannot unmarshal YML to map[string]interface{}: %w", err)
	}

	result, err := json.Marshal(yamlData)

	if err != nil {
		return nil, fmt.Errorf("cannot marshal map[string]interface{} to JSON: %w", err)
	}

	return result, nil
}
