package cmd

import (
	"errors"
	"log"
	"time"

	"github.com/servletcloud/Andmerada/internal/osutil"
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
			var projectDir string
			if len(args) > 0 {
				projectDir = args[0]
			} else {
				projectDir = osutil.GetwdOrPanic()
			}

			if err := InitializeProject(projectDir); err != nil {
				if errors.Is(err, ErrConfigFileAlreadyExists) {
					log.Fatalln(resources.MsgErrProjectExists())
				}
				log.Panic(err)
			}

			log.Printf("Project initialized successfully in %v", projectDir)
			log.Println(resources.MsgInitCompleted())
		},
	}
}

func createMigrationCmd() *cobra.Command {
	description := resources.LoadCrMigrationDescription()

	//nolint:exhaustruct
	return &cobra.Command{
		Use:   description.Use,
		Short: description.Short,
		Long:  description.Long,
		Args:  cobra.ExactArgs(1),
		Run: func(_ *cobra.Command, args []string) {
			currentDir := osutil.GetwdOrPanic()
			name := args[0]
			now := time.Now().UTC()

			result, err := CreateMigration(currentDir, name, now)

			if err != nil {
				if errors.Is(err, ErrMigrationAlreadyExists) {
					log.Fatalln(err)
				}
				log.Panic(err)
			}

			if !result.Latest {
				log.Println(resources.MsgMigrationNotLatest())
			}

			log.Println(resources.MsgMigrationCreated(result.BaseDir))
		},
		Example: `andmerada create-migration "Add users table"`,
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
	rootCmd.AddCommand(createMigrationCmd())

	return rootCmd
}
