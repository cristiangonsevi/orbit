package cmd

import (
	"fmt"

	"github.com/cristiangonsevi/orbit/internal/config"
	"github.com/cristiangonsevi/orbit/internal/ui"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate the configuration file",
	Long: `Validates the orbit configuration file and reports any warnings.

Unlike errors, warnings do not prevent orbit from running projects, but they
may indicate misconfigurations such as missing 'after' sections or empty
local sections.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ui.SetQuietMode(quiet)
		ui.Header("Validating Configuration")

		spinner := ui.NewSpinner("Loading configuration...")
		spinner.Start()

		cfg, err := config.LoadConfig(configFile)
		spinner.Stop()

		if err != nil {
			ui.Error(fmt.Sprintf("Config file has errors: %v", err))
			return err
		}

		warnings := cfg.ValidateDeep()

		fmt.Println()

		if len(warnings) == 0 {
			ui.Success(fmt.Sprintf("Config file %s is valid", configFile))
			if configFile == "" {
				ui.Info(fmt.Sprintf("Location: %s", config.DefaultConfigPath()))
			}
			return nil
		}

		for _, w := range warnings {
			msg := fmt.Sprintf("Project %q: %s", w.Project, w.Message)
			ui.Warning(msg)
		}

		fmt.Println()
		ui.Info("Run orbit run without --dry-run to execute")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}
