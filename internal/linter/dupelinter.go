package linter

import "fmt"

type DupeLinter struct {
	idToNames map[uint64][]string
}

func NewDupeLinter() DupeLinter {
	return DupeLinter{idToNames: make(map[uint64][]string)}
}

func (linter *DupeLinter) LintSource(id uint64, name string) {
	linter.idToNames[id] = append(linter.idToNames[id], name)
}

func (linter *DupeLinter) Report(report *Report) {
	for id, names := range linter.idToNames {
		if len(names) <= 1 {
			continue
		}

		message := fmt.Sprint(
			fmt.Sprintf("Duplicate migration ID: %v\n", id),
			"Ensure each migration has a unique timestamp-based ID.",
		)

		report.AddError(message, names...)
	}
}
