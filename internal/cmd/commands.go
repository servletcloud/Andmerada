package cmd

import (
	"log"
	"os"

	"github.com/spf13/cobra"
)

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
			initializeProject(targetDir)
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
