package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func AssertFileContains(t *testing.T, path string, expected string) {
	t.Helper()

	content := ReadFileToString(t, path)

	assert.Contains(t, content, expected)
}

func AssertPlaceholdersResolved(t *testing.T, content string) {
	t.Helper()

	assert.NotContains(t, content, "{{")
	assert.NotContains(t, content, "}}")
}
