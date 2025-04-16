package linter_test

import (
	"testing"

	"github.com/servletcloud/Andmerada/internal/linter"
	"github.com/stretchr/testify/assert"
)

func TestCountLint(t *testing.T) {
	t.Parallel()

	t.Run("reports a warning if there is no migrations", func(t *testing.T) {
		t.Parallel()

		report := linter.Report{} //nolint:exhaustruct
		linter := &linter.CountLinter{}

		linter.Report(&report)

		assert.Empty(t, report.Errors)
		assert.Len(t, report.Warnings, 1)
		assertContainsError(t, report.Warnings, "No migration files found. Create with:")
	})

	t.Run("does not report warning if there are migrations", func(t *testing.T) {
		t.Parallel()

		report := linter.Report{} //nolint:exhaustruct
		linter := &linter.CountLinter{}

		linter.LintSource()
		linter.Report(&report)

		assert.Empty(t, report.Errors)
		assert.Empty(t, report.Warnings)
	})
}
