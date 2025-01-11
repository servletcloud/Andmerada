package linters

import "fmt"

type CountLinter struct {
	Reporter
	hasMigrations bool
}

func (linter *CountLinter) LintSource() {
	linter.hasMigrations = true
}

func (linter *CountLinter) Report() {
	if linter.hasMigrations {
		return
	}

	message := fmt.Sprint(
		"No migration files found. Create with:\n",
		`andmerada create-migration "Add users table"`,
	)
	linter.AddWarning(message)
}
