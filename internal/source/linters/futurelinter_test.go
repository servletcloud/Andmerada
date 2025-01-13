package linters_test

import (
	"testing"

	"github.com/servletcloud/Andmerada/internal/source/linters"
	"github.com/stretchr/testify/assert"
)

func TestFutureLinter(t *testing.T) {
	t.Parallel()

	t.Run("no warning if no migrations in the future", func(t *testing.T) {
		t.Parallel()

		report := new(TestLintReport)
		linter := linters.FutureLinter{Reporter: report, Threshold: 3}

		linter.LintSource(1, "add users table")
		linter.LintSource(2, "add profiles table")
		linter.LintSource(3, "drop old_users materialized view")

		linter.Report()

		assert.Empty(t, report.errors)
		assert.Empty(t, report.warnings)
	})

	t.Run("returns a warning if there is a migration in the future", func(t *testing.T) {
		t.Parallel()

		report := new(TestLintReport)
		linter := linters.FutureLinter{Reporter: report, Threshold: 2}

		linter.LintSource(1, "add users table")
		linter.LintSource(2, "add profiles table")
		linter.LintSource(3, "drop old_users materialized view")

		linter.Report()

		assert.Empty(t, report.errors)
		assert.Len(t, report.warnings, 1)
		assert.Contains(t, report.warnings[0], "There are migrations with timestamps in the future")
	})
}
