package cmd

import (
	"errors"
	"log"

	"github.com/servletcloud/Andmerada/internal/cmd/descriptions"
	"github.com/servletcloud/Andmerada/internal/osutil"
	"github.com/servletcloud/Andmerada/internal/project"
	"github.com/servletcloud/Andmerada/internal/resources"
	"github.com/spf13/cobra"
)

func initProjectCommand() *cobra.Command {
	description := descriptions.InitDescription()

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

			if err := project.Initialize(projectDir); err != nil {
				if errors.Is(err, project.ErrConfigFileAlreadyExists) {
					log.Fatalln(resources.MsgErrProjectExists())
				} else if errors.Is(err, project.ErrNameExceeds255) {
					log.Fatalln("Error: Project name cannot exceed 255 characters in length")
				}
				log.Panic(err)
			}

			log.Printf("Project initialized successfully in %v", projectDir)
			log.Println(resources.MsgInitCompleted())
		},
	}
}
