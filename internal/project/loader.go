package project

import (
	"fmt"
	"path/filepath"

	"github.com/servletcloud/Andmerada/internal/schema"
	"github.com/servletcloud/Andmerada/internal/ymlutil"
)

func load(dir string) (Project, error) {
	configFileName := filepath.Join(dir, rootConfigFilename)

	var configuration Configuration

	if err := ymlutil.LoadFromFile(configFileName, schema.GetAndmeradaSchema(), &configuration); err != nil {
		return Project{}, fmt.Errorf("failed to load project configuration file %q: %w", configFileName, err)
	}

	return Project{Dir: dir, Configuration: configuration}, nil
}
