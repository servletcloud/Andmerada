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

	if err := applier.ApplyPending(cmd.Context()); err != nil {
		if !m.tryPrettyPrintError(err) {
			log.Panic(err)
		}
	}
}

func (m *migrateCmdRunner) tryPrettyPrintError(err error) bool {
	return m.tryPrettyPrintCancellation(err) ||
		m.tryPrettyPrintConnectError(err) ||
		m.tryPrettyPrintDDLError(err)
}

func (m *migrateCmdRunner) tryPrettyPrintCancellation(err error) bool {
	if errors.Is(err, context.Canceled) {
		log.Printf("Execution was cancelled by the user or due to a system timeout: %v", err)

		return true
	}

	if errors.Is(err, context.DeadlineExceeded) {
		log.Printf("Execution timed out before completing: %v", err)

		return true
	}

	return false
}

func (m *migrateCmdRunner) tryPrettyPrintConnectError(err error) bool {
	var connectErr *migrator.PostgresConnectError

	if !errors.As(err, &connectErr) {
		return false
	}

	var parseConfigErr *pgconn.ParseConfigError
	if errors.As(err, &parseConfigErr) {
		helpURL := "https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING"
		log.Printf("Invalid database URL: %v\n\nRead more at %v", parseConfigErr, helpURL)
	} else {
		log.Printf("Failed to connect to the database: %v", connectErr)
	}

	return true
}

func (m *migrateCmdRunner) tryPrettyPrintDDLError(err error) bool {
	var ddlErr *migrator.CreateDDLFailedError

	if !errors.As(err, &ddlErr) {
		return false
	}

	var pgError *pgconn.PgError
	if errors.As(ddlErr, &pgError) {
		log.Println(prettyPrintPgErr(pgError, ddlErr.SQL))
	} else {
		log.Println(ddlErr.Error())
	}

	log.Println("Failed to create auxiliary tables for managing migrations.")
	log.Println("Run 'andmerada show-ddl' to view the DDL SQL if you need to execute it manually.")

	return true
}
