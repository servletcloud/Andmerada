package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"

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

	if errors.Is(err, ymlutil.ErrSchemaValidation) {
		log.Fatalf("Schema validation failed for andmerada.yml: %v", err)
	}

	var yamlError *yaml.TypeError
	if errors.As(err, &yamlError) {
		log.Fatalf("Cannot parse andmerada.yml: %v", yamlError)
	}

	panic(fmt.Sprintf("Cannot read or parse the project: %v", err))
}
