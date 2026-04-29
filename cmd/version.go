package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd represents the `orbit version` command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display the current version",
	Long: `Print the version number of the orbit CLI.

This is also available via the --version flag.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("orbit version %s\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
