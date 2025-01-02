package cmd

import (
	"github.com/servletcloud/Andmerada/internal/resources"
	"github.com/spf13/cobra"
)

func lintCommand() *cobra.Command {
	description := resources.LoadLintCommandDescription()

	//nolint:exhaustruct
	return &cobra.Command{
		Use:   description.Use,
		Short: description.Short,
		Long:  description.Long,
		Run: func(_ *cobra.Command, _ []string) {

		},
	}
}
