package ymlutil

import (
	"fmt"
	"os"
	"strings"

	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
)

type ValidationError struct {
	validationResult *gojsonschema.Result
}

func (e *ValidationError) Error() string {
	return fmt.Sprintln("YML validation failed. Details:", e.Details())
}

func (e *ValidationError) Details() string {
	var stringBuilder strings.Builder

	for _, desc := range e.validationResult.Errors() {
		stringBuilder.WriteString(fmt.Sprintf("- %s\n", desc.String()))
	}

	return strings.TrimSpace(stringBuilder.String())
}

func NewValidationError(result *gojsonschema.Result) *ValidationError {
	return &ValidationError{validationResult: result}
}

func LoadFromFile(path string, schema string, out any) error {
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
		return fmt.Errorf("cannot validate schema in %q : %w", path, NewValidationError(result))
	}

	return nil
}
