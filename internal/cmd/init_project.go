package cmd

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/servletcloud/Andmerada/internal/resources"
)

const (
	rootConfigFilename string = "andmerada.yml"

	msgProjectAlreadyExists string = `Error: Project initialization failed.
The file 'andmerada.yml' already exists in the specified directory.

Suggestion:
- If this is an existing Andmerada project, you can start by running commands like:
  andmerada create-migration "Add users table"

- If you want to reinitialize the project, please remove or back up the existing 'andmerada.yml' file and try again.`

	msgNextSteps string = `Next Steps:
1. Configure your project:
   Edit 'andmerada.yml' to update the project name, migrations table name, and other settings.

2. Create your first migration:
   andmerada create-migration "Add users table"

3. Run your first migration:
   andmerada migrate`
)

func initializeProject(targetDir string) {
	if err := os.MkdirAll(targetDir, os.ModePerm); err != nil {
		log.Fatalf("failed to create directory %s: %v", targetDir, err)
	}

	configFilePath := filepath.Join(targetDir, rootConfigFilename)
	file, err := os.OpenFile(configFilePath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, os.ModePerm)

	if err != nil {
		if os.IsExist(err) {
			log.Fatal(msgProjectAlreadyExists)
		}

		log.Fatalf("Failed to create file: %v\n", err)
	}

	projectName := filepath.Base(filepath.Dir(configFilePath))
	content := strings.ReplaceAll(resources.TemplateAndmeradaYml(), "{{project_name}}", projectName)

	if _, err = file.WriteString(content); err != nil {
		log.Fatalf("Failed to create file: %v\n", err)
	}

	log.Printf("Project '%v' initialized successfully in %v", projectName, targetDir)
	log.Println(msgNextSteps)

	defer file.Close()
}
