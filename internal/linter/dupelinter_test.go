package linter_test

import (
	"testing"

	"github.com/servletcloud/Andmerada/internal/linter"
	"github.com/stretchr/testify/assert"
)

func TestDupeLinter(t *testing.T) {
	t.Parallel()

	t.Run("no errors if migrations have unique IDs", func(t *testing.T) {
		t.Parallel()

		report := linter.Report{} //nolint:exhaustruct
		linter := linter.NewDupeLinter()

		linter.LintSource(1, "add users table")
		linter.LintSource(2, "add profiles table")
		linter.LintSource(3, "drop old_users materialized view")

		linter.Report(&report)

		assert.Empty(t, report.Errors)
		assert.Empty(t, report.Warnings)
	})

	t.Run("returns an error if there is a duplicate migration ID", func(t *testing.T) {
		t.Parallel()

		report := linter.Report{} //nolint:exhaustruct
		linter := linter.NewDupeLinter()

		linter.LintSource(1, "add users table")
		linter.LintSource(2, "add profiles table")
		linter.LintSource(2, "drop old_users materialized view")

		linter.Report(&report)

		assert.Len(t, report.Errors, 1)
		assertContainsError(t, report.Errors, "Duplicate migration ID")
		assert.Empty(t, report.Warnings)
	})
}
