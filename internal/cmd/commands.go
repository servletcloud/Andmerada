package cmd

import (
	"errors"
	"log"
	"os"

	"github.com/spf13/cobra"
)

const msgProjectAlreadyExists string = `Error: Project initialization failed.
The file 'andmerada.yml' already exists in the specified directory.

Suggestion:
- If this is an existing Andmerada project, you can start by running commands like:
 andmerada create-migration "Add users table"

- If you want to reinitialize the project, please remove or back up the existing 'andmerada.yml' file and try again.`

const msgInitCompleted string = `Next Steps:
1. Configure your project:
  Edit 'andmerada.yml' to update the project name, migrations table name, and other settings.

2. Create your first migration:
  andmerada create-migration "Add users table"

3. Run your first migration:
  andmerada migrate`

func initCommand() *cobra.Command {
	//nolint:exhaustruct
	return &cobra.Command{
		Use:   "init [directory]",
		Short: "Initialize a new Andmerada migration project",
		Long: `The 'andmerada init' command sets up a new migration project.
If no directory is provided, the current directory will be used.
It pre-creates the specified directory (including nested folders) if it does not
already exist and generates an 'andmerada.yml' configuration file.`,
		Args: cobra.MaximumNArgs(1),
		Run: func(_ *cobra.Command, args []string) {
			var targetDir string
			if len(args) > 0 {
				targetDir = args[0]
			} else {
				currentDir, err := os.Getwd()
				if err != nil {
					log.Fatalf("Error getting current directory: %v\n", err)
				}
				targetDir = currentDir
			}

			if err := initializeProject(targetDir); err != nil {
				if errors.Is(err, errConfigFileAlreadyExists) {
					log.Fatalln(msgProjectAlreadyExists)
				}
				log.Panic(err)
			}

			log.Printf("Project initialized successfully in %v", targetDir)
			log.Println(msgInitCompleted)
		},
	}
}

func RootCommand() *cobra.Command {
	//nolint:exhaustruct
	rootCmd := &cobra.Command{
		Use: "andmerada",
		Annotations: map[string]string{
			cobra.CommandDisplayNameAnnotation: "andmerada",
		},
		Version: "0.0.1",
	}
	rootCmd.AddCommand(initCommand())

	return rootCmd
}
