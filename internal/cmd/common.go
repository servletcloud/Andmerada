package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/servletcloud/Andmerada/internal/project"
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

	var yamlError *yaml.TypeError
	if errors.As(err, &yamlError) {
		log.Fatalf("Cannot parse andmerada.yml: %v", yamlError)
	}

	panic(fmt.Sprintf("Cannot read or parse the project: %v", err))
}
