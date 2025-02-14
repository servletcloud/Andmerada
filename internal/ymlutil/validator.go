package ymlutil

import (
	"encoding/json"
	"fmt"

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

func ymlToJSON(yml []byte) ([]byte, error) {
	var yamlData map[string]any

	if err := yaml.Unmarshal(yml, &yamlData); err != nil {
		return nil, fmt.Errorf("cannot unmarshal YML to map[string]any: %w", err)
	}

	result, err := json.Marshal(yamlData)

	if err != nil {
		return nil, fmt.Errorf("cannot marshal map[string]any to JSON: %w", err)
	}

	return result, nil
}
