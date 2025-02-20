package cmd

import (
	"context"
	"errors"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/servletcloud/Andmerada/internal/cmd/descriptions"
	"github.com/servletcloud/Andmerada/internal/migrator"
	"github.com/servletcloud/Andmerada/internal/osutil"
	"github.com/spf13/cobra"
)

func migrateCommand() *cobra.Command {
	description := descriptions.MigrateDescription()
	migrate := migrateCmdRunner{}

	//nolint:exhaustruct
	command := &cobra.Command{
		Use:   description.Use,
		Short: description.Short,
		Long:  description.Long,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, _ []string) {
			migrate.Run(cmd)
		},
		Example: `andmerada migrate --database-url=postgres://postgres:secret@localhost:5432/mydatabase?sslmode=disable`,
	}

	command.Flags().String(
		"database-url",
		os.Getenv("DATABASE_URL"),
		"Database connection URL (defaults to DATABASE_URL environment variable)",
	)

	return command
}

type migrateCmdRunner struct {
}

func (m *migrateCmdRunner) Run(cmd *cobra.Command) {
	databaseURL, err := cmd.Flags().GetString("database-url")
	if err != nil {
		log.Fatalf("Failed to retrieve the --database-url flag: %v", err)
	}

	if len(databaseURL) == 0 {
		log.Fatalf("Database URL is missing. Set the DATABASE_URL environment variable or specify it " +
			"using the --database-url flag.")
	}

	applier := &migrator.Applier{
		Project:     mustLoadProject(osutil.GetwdOrPanic()),
		DatabaseURL: databaseURL,
	}
	report := migrator.Report{SourcesOnDisk: 0}

	if err := applier.ApplyPending(cmd.Context(), &report); err != nil {
		m.printError(err)
	} else {
		m.printReport(&report)
	}
}

func (m *migrateCmdRunner) printError(err error) {
	if m.isCancellationError(err) {
		m.printCancellationError(err)

		return
	}

	var migratorErr *migrator.MigrateError

	if !errors.As(err, &migratorErr) {
		log.Panic(err)
	}

	switch migratorErr.ErrType {
	case migrator.ErrTypeDBConnect:
		m.printConnectError(migratorErr)
	case migrator.ErrTypeCreateDDL:
		m.printDDLError(migratorErr)
	case migrator.ErrTypeListMigrationsOnDisk:
		log.Printf("Failed to list migrations on disk: %v", migratorErr)
	case migrator.ErrTypeScanAppliedMigrations:
		log.Printf("Failed to scan applied migrations: %v", m.pgErrorToPrettyString(migratorErr))
	default:
		log.Panic(migratorErr)
	}
}

func (m *migrateCmdRunner) printCancellationError(err error) {
	switch {
	case errors.Is(err, context.Canceled):
		log.Printf("Execution was cancelled by the user or due to a system timeout: %v", err)
	case errors.Is(err, context.DeadlineExceeded):
		log.Printf("Execution timed out before completing: %v", err)
	}
}

func (m *migrateCmdRunner) isCancellationError(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

func (m *migrateCmdRunner) printConnectError(err *migrator.MigrateError) {
	var parseConfigErr *pgconn.ParseConfigError

	if errors.As(err, &parseConfigErr) {
		helpURL := "https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING"
		log.Printf("Invalid database URL: %v\n\nRead more at %v", parseConfigErr, helpURL)
	} else {
		log.Printf("Failed to connect to the database: %v", err)
	}
}

func (m *migrateCmdRunner) printDDLError(err *migrator.MigrateError) {
	log.Println(m.pgErrorToPrettyString(err))
	log.Println("Failed to create auxiliary tables for managing migrations.")
	log.Println("Run 'andmerada show-ddl' to view the DDL SQL if you need to execute it manually.")
}

func (m *migrateCmdRunner) pgErrorToPrettyString(err *migrator.MigrateError) string {
	var pgError *pgconn.PgError

	if errors.As(err, &pgError) {
		return (&pgErrorTranslator{}).prettyPrint(pgError, err.SQL)
	}

	return err.Error()
}

func (m *migrateCmdRunner) printReport(report *migrator.Report) {
	if report.SourcesOnDisk == 0 {
		help := `andmerada create-migration "Add users table"`
		log.Println("No migration files found. To create one, run:\n" + help)
	}
}
