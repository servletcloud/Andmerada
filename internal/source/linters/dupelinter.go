package linters

import "fmt"

type DupeLinter struct {
	Reporter
	idToNames map[uint64][]string
}

func (linter *DupeLinter) LintSource(id uint64, name string) {
	linter.idToNames[id] = append(linter.idToNames[id], name)
}

func (linter *DupeLinter) Report() {
	for id, names := range linter.idToNames { //nolint:varnamelen
		if len(names) <= 1 {
			continue
		}

		message := fmt.Sprint(
			fmt.Sprintf("Duplicate migration ID: %v\n", id),
			"Ensure each migration has a unique timestamp-based ID.",
		)

		linter.AddError(message, names...)
	}
}
