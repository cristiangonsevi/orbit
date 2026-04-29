package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// configFile overrides the default config file path
	configFile string
	// verbose enables detailed output
	verbose bool
	// Version is set at build time via ldflags
	Version = "0.1.0"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "orbit",
	Short: "Remote CLI automation tool",
	Long: `Orbit is a CLI tool for automating remote server workflows.

It combines SSH access, file transfers, and local/remote command execution
into configurable, reusable projects.

Complete documentation is available at:
  https://github.com/cristiangonsevi/orbit`,
	Version: Version,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "Path to config file (default: ~/.config/orbit/config.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output for debugging")
}
