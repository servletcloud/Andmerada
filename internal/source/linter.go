package source

import (
	"fmt"
	"path/filepath"

	"github.com/servletcloud/Andmerada/internal/source/linters"
)

type linter struct {
	LintConfiguration
	report *LintReport
}

func (linter *linter) lint() error {
	duplicatesLinter := linters.NewDupeLinter(linter)
	defer duplicatesLinter.Report()

	futureLinter := linter.newFutureLinter()
	defer futureLinter.Report()

	countLinter := &linters.CountLinter{Reporter: linter}
	defer countLinter.Report()

	configurationLinter := &linters.ConfigLinter{Reporter: linter, ProjectDir: linter.ProjectDir}
	upSQLLinter := linter.newUpSQLLinter()
	downSQLLinter := linter.newDownSQLLinter()

	return scanAll(linter.ProjectDir, func(id MigrationID, name string) {
		duplicatesLinter.LintSource(id.asUint64(), name)
		futureLinter.LintSource(id.asUint64(), name)
		countLinter.LintSource()

		configuration := new(Configuration)

		if !configurationLinter.Lint(filepath.Join(name, MigrationYmlFilename), &configuration) {
			return
		}

		upSQLLinter.Lint(filepath.Join(name, configuration.Up.File))

		if !configuration.Down.Block {
			downSQLLinter.Lint(filepath.Join(name, configuration.Down.File))
		}
	})
}

func (linter *linter) newUpSQLLinter() linters.SQLLinter {
	return linters.SQLLinter{
		Reporter:            linter,
		ProjectDir:          linter.ProjectDir,
		MaxSQLFileSize:      linter.MaxSQLFileSize,
		CreatedFromTemplate: []byte(linter.UpSQLTemplate),
		ErrEmptyMsg:         "Migration file appears to be empty.",
		ErrUntouchedMsg:     "The migration file appears to be untouched since its creation.",
	}
}

func (linter *linter) newDownSQLLinter() linters.SQLLinter {
	errEmptyMsg := fmt.Sprint(
		"The migration rollback file appears to be empty. ",
		"Consider adding a comment or marking the rollback as blocked.",
	)

	return linters.SQLLinter{
		Reporter:            linter,
		ProjectDir:          linter.ProjectDir,
		MaxSQLFileSize:      linter.MaxSQLFileSize,
		CreatedFromTemplate: []byte(linter.DownSQLTemplate),
		ErrEmptyMsg:         errEmptyMsg,
		ErrUntouchedMsg:     "The migration rollback file appears to be untouched since its creation.",
	}
}

func (linter *linter) newFutureLinter() *linters.FutureLinter {
	return &linters.FutureLinter{
		Reporter:  linter,
		Threshold: newIDFromTime(linter.NowUTC).asUint64(),
	}
}

func (linter *linter) AddError(title string, files ...string) {
	linter.report.Errors = append(linter.report.Errors, LintError{Title: title, Files: files})
}

func (linter *linter) AddWarning(title string, files ...string) {
	linter.report.Warings = append(linter.report.Warings, LintError{Title: title, Files: files})
}
