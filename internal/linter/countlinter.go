package linter

import "fmt"

type CountLinter struct {
	hasMigrations bool
}

func (linter *CountLinter) LintSource() {
	linter.hasMigrations = true
}

func (linter *CountLinter) Report(report *Report) {
	if linter.hasMigrations {
		return
	}

	message := fmt.Sprint(
		"No migration files found. Create with:\n",
		`andmerada create-migration "Add users table"`,
	)
	report.AddWarning(message)
}
