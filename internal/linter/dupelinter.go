package linter

import (
	"fmt"

	"github.com/servletcloud/Andmerada/internal/source"
)

type DupeLinter struct {
	idToNames map[source.ID][]string
}

func NewDupeLinter() DupeLinter {
	return DupeLinter{idToNames: make(map[source.ID][]string)}
}

func (linter *DupeLinter) LintSource(id source.ID, name string) {
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
