package linter

import (
	"fmt"

	"github.com/servletcloud/Andmerada/internal/source"
)

type FutureLinter struct {
	Threshold       source.ID
	sourcesInFuture []string
}

func (linter *linter) newFutureLinter(threshold source.ID) *FutureLinter {
	return &FutureLinter{Threshold: threshold, sourcesInFuture: nil}
}

func (linter *FutureLinter) LintSource(id source.ID, name string) {
	if id <= linter.Threshold {
		return
	}

	linter.sourcesInFuture = append(linter.sourcesInFuture, name)
}

func (linter *FutureLinter) Report(report *Report) {
	if len(linter.sourcesInFuture) == 0 {
		return
	}

	message := fmt.Sprint(
		"There are migrations with timestamps in the future.\n",
		"These migrations are pending unless already applied, regardless of their timestamps",
	)
	report.AddWarning(message, linter.sourcesInFuture...)
}
