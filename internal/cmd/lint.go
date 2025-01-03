package cmd

import (
	"log"
	"os"
	"slices"
	"sort"
	"strings"

	"github.com/servletcloud/Andmerada/internal/osutil"
	"github.com/servletcloud/Andmerada/internal/resources"
	"github.com/servletcloud/Andmerada/internal/source"
	"github.com/spf13/cobra"
)

const (
	exitCodeLintErrors = 1
)

func lintCommand() *cobra.Command {
	description := resources.LoadLintCommandDescription()

	//nolint:exhaustruct
	return &cobra.Command{
		Use:   description.Use,
		Short: description.Short,
		Long:  description.Long,
		Run: func(_ *cobra.Command, _ []string) {
			currentDir := osutil.GetwdOrPanic()

			ensureProjectInitialized(currentDir)

			log.Println("Validating the migration files, please, wait...")

			report := source.LintReport{}
			if err := source.Lint(currentDir, &report); err != nil {
				log.Panic(err)
			}

			printLintReport(&report)

			if len(report.Errors) > 0 {
				os.Exit(exitCodeLintErrors)
			}
		},
	}
}

func printLintReport(report *source.LintReport) {
	if len(report.Errors) == 0 {
		log.Println("✔️ All checks passed. No issues found.")

		return
	}

	criticalCount := len(report.Errors)
	warningCount := len(report.Warings)

	log.Println("Lint Report:")

	if criticalCount > 0 {
		log.Printf("%d Critical Errors:\n", criticalCount)

		sortLintErrorsByIDAsc(report.Errors)

		for _, err := range report.Errors {
			printLintError(err)
		}
	}

	if warningCount > 0 {
		log.Printf("%d Warnings:\n", warningCount)

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
	log.Printf("  - %s\n", err.Title)

	if len(err.Details) > 0 {
		log.Printf("    Details: %s\n", err.Details)
	}

	if len(err.Files) > 0 {
		skip := false

		if len(err.Files) == 1 {
			file := err.Files[0]
			skip = strings.Contains(err.Details, file) || strings.Contains(err.Title, file)
		}

		if !skip {
			log.Printf("    Affected Files: %s\n", err.Files)
		}
	}

	log.Println()
}

func sortLintErrorsByIDAsc(errors []source.LintError) {
	sort.Slice(errors, func(i int, j int) bool {
		err1 := errors[i]
		err2 := errors[j]

		slices.Sort(err1.Files)
		slices.Sort(err2.Files)

		return strings.Compare(err1.Files[0], err2.Files[0]) <= 0
	})
}
