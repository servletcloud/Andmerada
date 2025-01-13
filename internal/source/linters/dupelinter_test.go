package linters_test

import (
	"testing"

	"github.com/servletcloud/Andmerada/internal/source/linters"
	"github.com/stretchr/testify/assert"
)

func TestDupeLinter(t *testing.T) {
	t.Parallel()

	t.Run("no errors if migrations have unique IDs", func(t *testing.T) {
		t.Parallel()

		report := new(TestLintReport)
		linter := linters.NewDupeLinter(report)

		linter.LintSource(1, "add users table")
		linter.LintSource(2, "add profiles table")
		linter.LintSource(3, "drop old_users materialized view")

		linter.Report()

		assert.Empty(t, report.errors)
		assert.Empty(t, report.warnings)
	})

	t.Run("returns an error if there is a duplicate migration ID", func(t *testing.T) {
		t.Parallel()

		report := new(TestLintReport)
		linter := linters.NewDupeLinter(report)

		linter.LintSource(1, "add users table")
		linter.LintSource(2, "add profiles table")
		linter.LintSource(2, "drop old_users materialized view")

		linter.Report()

		assert.Len(t, report.errors, 1)
		assert.Contains(t, report.errors[0], "Duplicate migration ID")
		assert.Empty(t, report.warnings)
	})
}
