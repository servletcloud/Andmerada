package cmd

import (
	"github.com/spf13/cobra"
)

func RootCommand() *cobra.Command {
	//nolint:exhaustruct
	rootCmd := &cobra.Command{
		Use: "andmerada",
		Annotations: map[string]string{
			cobra.CommandDisplayNameAnnotation: "andmerada",
		},
		Version: "0.0.1",
	}
	rootCmd.AddCommand(initProjectCommand())
	rootCmd.AddCommand(createMigrationCommand())

	return rootCmd
}
