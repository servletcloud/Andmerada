package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/servletcloud/Andmerada/internal/project"
	"github.com/servletcloud/Andmerada/internal/ymlutil"
	"gopkg.in/yaml.v3"
)

func ensureProjectInitialized(dir string) {
	_ = mustLoadProject(dir)
}

func mustLoadProject(dir string) project.Project {
	project, err := project.Load(dir)

	if err == nil {
		return project
	}

	if errors.Is(err, os.ErrNotExist) {
		log.Fatalf("Project is not initialized. Initialize with `andmerada init %v`", dir)
	}

	if schemaError := new(ymlutil.ValidationError); errors.As(err, &schemaError) {
		log.Fatalf("Schema validation failed for andmerada.yml:\n%v", schemaError.Details())
	}

	var yamlError *yaml.TypeError
	if errors.As(err, &yamlError) {
		log.Fatalf("Cannot parse andmerada.yml: %v", yamlError)
	}

	panic(fmt.Sprintf("Cannot read or parse the project: %v", err))
}

func prettyPrintPgErr(pgErr *pgconn.PgError, sql string) string {
	var sb strings.Builder

	translator := pgErrorTranslator{
		writeString: func(message string) {
			sb.WriteString(message)
		},
	}
	translator.prettyPrint(pgErr, sql)

	return sb.String()
}
