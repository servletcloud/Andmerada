package linter

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/servletcloud/Andmerada/internal/source"
)

type Configuration struct {
	ProjectDir      string
	MaxSQLFileSize  int64
	NowUTC          time.Time
	UpSQLTemplate   string
	DownSQLTemplate string
}

type LintError struct {
	Title string
	Files []string
}

func Run(conf Configuration, report *Report) error {
	linter := linter{Configuration: conf}

	return linter.lint(report)
}

type linter struct {
	Configuration
}

func (linter *linter) lint(report *Report) error {
	duplicatesLinter := NewDupeLinter()
	defer duplicatesLinter.Report(report)

	futureLinter := linter.newFutureLinter(source.NewIDFromTime(linter.NowUTC))
	defer futureLinter.Report(report)

	countLinter := &CountLinter{} //nolint:exhaustruct
	defer countLinter.Report(report)

	configurationLinter := &ConfigLinter{ProjectDir: linter.ProjectDir}
	upSQLLinter := linter.newUpSQLLinter()
	downSQLLinter := linter.newDownSQLLinter()

	return source.TraverseAll(linter.ProjectDir, func(id uint64, name string) { //nolint:wrapcheck
		duplicatesLinter.LintSource(id, name)
		futureLinter.LintSource(id, name)
		countLinter.LintSource()

		configPath := filepath.Join(name, source.MigrationYmlFilename)
		configuration := new(source.Configuration)

		if !configurationLinter.Lint(report, configPath, &configuration) {
			return
		}

		upSQLLinter.Lint(report, filepath.Join(name, configuration.Up.File))

		if !configuration.Down.Block {
			downSQLLinter.Lint(report, filepath.Join(name, configuration.Down.File))
		}
	})
}

func (linter *linter) newUpSQLLinter() SQLLinter {
	return SQLLinter{
		ProjectDir:          linter.ProjectDir,
		MaxSQLFileSize:      linter.MaxSQLFileSize,
		CreatedFromTemplate: []byte(linter.UpSQLTemplate),
		ErrEmptyMsg:         "Migration file appears to be empty.",
		ErrUntouchedMsg:     "The migration file appears to be untouched since its creation.",
	}
}

func (linter *linter) newDownSQLLinter() SQLLinter {
	errEmptyMsg := fmt.Sprint(
		"The migration rollback file appears to be empty. ",
		"Consider adding a comment or marking the rollback as blocked.",
	)

	return SQLLinter{
		ProjectDir:          linter.ProjectDir,
		MaxSQLFileSize:      linter.MaxSQLFileSize,
		CreatedFromTemplate: []byte(linter.DownSQLTemplate),
		ErrEmptyMsg:         errEmptyMsg,
		ErrUntouchedMsg:     "The migration rollback file appears to be untouched since its creation.",
	}
}
