package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the application version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s version %s\n", cfg.App.Name, cfg.App.Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
