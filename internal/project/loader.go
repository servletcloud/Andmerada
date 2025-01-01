package project

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func load(dir string) (Project, error) {
	configFileName := filepath.Join(dir, rootConfigFilename)

	configBytes, err := os.ReadFile(configFileName)
	if err != nil {
		return Project{}, fmt.Errorf("can not read project configuration %v: %w", configFileName, err)
	}

	var configuration Configuration
	if err = yaml.Unmarshal(configBytes, &configuration); err != nil {
		return Project{}, fmt.Errorf("can not parse project configuration %v: %w", configFileName, err)
	}

	return Project{Dir: dir, Configuration: configuration}, nil
}
