package linter_test

import (
	"testing"

	"github.com/servletcloud/Andmerada/internal/linter"
	"github.com/stretchr/testify/assert"
)

func TestFutureLinter(t *testing.T) {
	t.Parallel()

	t.Run("no warning if no migrations in the future", func(t *testing.T) {
		t.Parallel()

		report := linter.Report{} //nolint:exhaustruct
		linter := linter.FutureLinter{Threshold: 3}

		linter.LintSource(1, "add users table")
		linter.LintSource(2, "add profiles table")
		linter.LintSource(3, "drop old_users materialized view")

		linter.Report(&report)

		assert.Empty(t, report.Errors)
		assert.Empty(t, report.Warnings)
	})

	t.Run("returns a warning if there is a migration in the future", func(t *testing.T) {
		t.Parallel()

		report := linter.Report{} //nolint:exhaustruct
		linter := linter.FutureLinter{Threshold: 2}

		linter.LintSource(1, "add users table")
		linter.LintSource(2, "add profiles table")
		linter.LintSource(3, "drop old_users materialized view")

		linter.Report(&report)

		assert.Empty(t, report.Errors)
		assert.Len(t, report.Warnings, 1)
		assertContainsError(t, report.Warnings, "There are migrations with timestamps in the future")
	})
}
