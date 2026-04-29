package cmd

import (
	_ "embed"
	"fmt"

	"github.com/cristiangonsevi/orbit/internal/config"
	"github.com/spf13/cobra"
)

//go:embed templates/config.yaml
var defaultConfigTemplate string

// initCmd represents the `orbit init` command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new configuration file",
	Long: `Create a default YAML configuration file.

This command creates the config file at the default location:
  ~/.config/orbit/config.yaml

If the file already exists, it will print an error and exit.
Use --config to specify a custom path, but init always creates
the default config file for bootstrapping.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if verbose {
			fmt.Println("Initializing orbit configuration...")
		}

		path, err := config.InitConfig(defaultConfigTemplate)
		if err != nil {
			return fmt.Errorf("init failed: %w", err)
		}

		fmt.Printf("Configuration template created at: %s\n", path)
		fmt.Println("Edit this file to define your projects, then run:")
		fmt.Println("  orbit list          # list available projects")
		fmt.Println("  orbit run <name>    # run a project workflow")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
