package ymlutil

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

var (
	ErrSchemaValidation = errors.New("cannot validate schema")
)

func LoadFromFile(path string, schema string, out interface{}) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("cannot read file %q: %w", path, err)
	}

	if err = yaml.Unmarshal(content, out); err != nil {
		return fmt.Errorf("cannot parse YML file %q: %w", path, err)
	}

	result, err := validate(content, schema)
	if err != nil {
		return fmt.Errorf("cannot validate JSON schema in %q: %w", path, err)
	}

	if !result.Valid() {
		return fmt.Errorf("cannot validate schema in %q because %v: %w",
			path,
			fmtValidationErrors(result),
			ErrSchemaValidation,
		)
	}

	return nil
}
