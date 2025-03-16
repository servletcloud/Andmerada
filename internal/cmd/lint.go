package cmd

import (
	"log"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/dustin/go-humanize/english"
	"github.com/servletcloud/Andmerada/internal/cmd/descriptions"
	"github.com/servletcloud/Andmerada/internal/osutil"
	"github.com/servletcloud/Andmerada/internal/resources"
	"github.com/servletcloud/Andmerada/internal/source"
	"github.com/spf13/cobra"
)

const (
	exitCodeLintErrors = 1
)

func lintCommand() *cobra.Command {
	description := descriptions.LintDescription()

	return &cobra.Command{ //nolint:exhaustruct
		Use:   description.Use,
		Short: description.Short,
		Long:  description.Long,
		Run: func(_ *cobra.Command, _ []string) {
			currentDir := osutil.GetwdOrPanic()

			ensureProjectInitialized(currentDir)

			log.Println("Validating the migration files, please, wait...")
			log.Println()

			config := source.LintConfiguration{
				ProjectDir:      currentDir,
				MaxSQLFileSize:  MaxSQLFileSizeBytes,
				NowUTC:          time.Now().UTC(),
				UpSQLTemplate:   resources.TemplateUpSQL(),
				DownSQLTemplate: resources.TemplateDownSQL(),
			}
			report := new(source.LintReport)
			if err := source.Lint(config, report); err != nil {
				log.Panic(err)
			}

			printLintReport(report)

			if len(report.Errors) > 0 {
				os.Exit(exitCodeLintErrors)
			}
		},
	}
}

func printLintReport(report *source.LintReport) {
	if len(report.Errors) == 0 && len(report.Warings) == 0 {
		log.Println("✔️ All checks passed. No issues found.")

		return
	}

	criticalCount := len(report.Errors)
	warningCount := len(report.Warings)

	if criticalCount > 0 {
		log.Println("Critical Errors:")

		sortLintErrorsByIDAsc(report.Errors)

		for _, err := range report.Errors {
			printLintError(err)
		}
	}

	if warningCount > 0 {
		log.Println("Warnings:")

		sortLintErrorsByIDAsc(report.Warings)

		for _, err := range report.Warings {
			printLintError(err)
		}
	}

	log.Printf(" Summary: Critical Errors: %d, Warnings: %d", criticalCount, warningCount)

	if criticalCount > 0 {
		log.Println("  ⚠️ Critical errors detected. 'andmerada migrate' will fail to run these migrations.")
	}
}

func printLintError(err source.LintError) {
	for i, line := range strings.Split(err.Title, "\n") {
		if i == 0 {
			log.Printf("  - %s\n", line)
		} else {
			log.Printf("    %s\n", line)
		}
	}

	if len(err.Files) > 0 {
		log.Printf("    Affected %v: %v",
			english.PluralWord(len(err.Files), "file", "files"),
			english.WordSeries(err.Files, "and"),
		)
	}

	log.Println()
}

func sortLintErrorsByIDAsc(errors []source.LintError) {
	slices.SortFunc(errors, func(a, b source.LintError) int { //nolint:varnamelen
		if len(a.Files) == 0 {
			return 1
		}

		if len(b.Files) == 0 {
			return -1
		}

		slices.Sort(a.Files)
		slices.Sort(b.Files)

		return strings.Compare(a.Files[0], b.Files[0])
	})
}
