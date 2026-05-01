package cmd

import (
	_ "embed"
	"fmt"
	"time"

	"github.com/cristiangonsevi/orbit/internal/config"
	"github.com/cristiangonsevi/orbit/internal/ui"
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
		ui.SetQuietMode(quiet)
		ui.Header("Orbit Init")

		spinner := ui.NewSpinner("Creating configuration...")
		spinner.Start()

		// Small delay for visual effect
		time.Sleep(300 * time.Millisecond)

		path, err := config.InitConfig(defaultConfigTemplate)
		spinner.Stop()

		if err != nil {
			ui.Error(fmt.Sprintf("Failed to initialize: %v", err))
			return err
		}

		ui.Success(fmt.Sprintf("Configuration created at: %s", path))
		fmt.Println()
		ui.Info("Next steps:")
		fmt.Printf("  %s\n", ui.ColorBold("orbit list"))
		fmt.Printf("  %s %s\n", ui.ColorBold("orbit run"), ui.ColorDim("<project-name>"))
		fmt.Println()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
