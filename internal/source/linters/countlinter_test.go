package linters_test

import (
	"testing"

	"github.com/servletcloud/Andmerada/internal/source/linters"
	"github.com/stretchr/testify/assert"
)

func TestCountLint(t *testing.T) {
	t.Parallel()

	t.Run("reports a warning if there is no migrations", func(t *testing.T) {
		t.Parallel()

		report := new(TestLintReport)
		linter := &linters.CountLinter{Reporter: report}

		linter.Report()

		assert.Empty(t, report.errors)
		assert.Len(t, report.warnings, 1)
		assert.Contains(t, report.warnings[0], "No migration files found. Create with:")
	})

	t.Run("does not report warning if there are migrations", func(t *testing.T) {
		t.Parallel()

		report := new(TestLintReport)
		linter := &linters.CountLinter{Reporter: report}

		linter.LintSource()
		linter.Report()

		assert.Empty(t, report.errors)
		assert.Empty(t, report.warnings)
	})
}
