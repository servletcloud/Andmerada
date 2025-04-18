package cmd

import (
	"errors"
	"log"

	"github.com/servletcloud/Andmerada/internal/cmd/descriptions"
	"github.com/servletcloud/Andmerada/internal/osutil"
	"github.com/servletcloud/Andmerada/internal/resources"
	"github.com/servletcloud/Andmerada/internal/source"
	"github.com/spf13/cobra"
)

func createMigrationCommand() *cobra.Command {
	description := descriptions.CrMigrationDescription()

	//nolint:exhaustruct
	return &cobra.Command{
		Use:   description.Use,
		Short: description.Short,
		Long:  description.Long,
		Args:  cobra.ExactArgs(1),
		Run: func(_ *cobra.Command, args []string) {
			currentDir := osutil.GetwdOrPanic()
			name := args[0]
			id := source.NewIDFromNow()

			ensureProjectInitialized(currentDir)

			result, err := source.Create(currentDir, name, id)

			if err != nil {
				if errors.Is(err, source.ErrSourceAlreadyExists) {
					log.Fatalln(err)
				} else if errors.Is(err, source.ErrNameExceeds255) {
					log.Fatalln("Error: Migration name cannot exceed 255 characters in length")
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
