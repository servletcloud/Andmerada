package cmd

import (
	"errors"
	"log"
	"os"

	"github.com/servletcloud/Andmerada/internal/resources"
	"github.com/spf13/cobra"
)

func initCommand() *cobra.Command {
	description := resources.LoadInitCommandDescription()

	//nolint:exhaustruct
	return &cobra.Command{
		Use:   description.Use,
		Short: description.Short,
		Long:  description.Long,
		Args:  cobra.MaximumNArgs(1),
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
					log.Fatalln(resources.MsgErrProjectExists())
				}
				log.Panic(err)
			}

			log.Printf("Project initialized successfully in %v", targetDir)
			log.Println(resources.MsgInitCompleted())
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
