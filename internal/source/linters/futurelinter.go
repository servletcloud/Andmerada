package linters

import "fmt"

type FutureLinter struct {
	Reporter
	Threshold       uint64
	sourcesInFuture []string
}

func (linter *FutureLinter) LintSource(id uint64, name string) {
	if id <= linter.Threshold {
		return
	}

	linter.sourcesInFuture = append(linter.sourcesInFuture, name)
}

func (linter *FutureLinter) Report() {
	if len(linter.sourcesInFuture) == 0 {
		return
	}

	message := fmt.Sprint(
		"There are migrations with timestamps in the future.\n",
		"These migrations are pending unless already applied, regardless of their timestamps",
	)
	linter.AddWarning(message, linter.sourcesInFuture...)
}
