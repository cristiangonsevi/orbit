package cmd

import (
	"fmt"
	"sort"

	"github.com/cristiangonsevi/orbit/internal/config"
	"github.com/cristiangonsevi/orbit/internal/ui"
	"github.com/spf13/cobra"
)

// getProjectNames returns all project names from the config file (reused from run.go)
func getProjectNamesFromConfig() ([]string, error) {
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(cfg.Projects))
	for name := range cfg.Projects {
		names = append(names, name)
	}
	return names, nil
}

// listCmd represents the `orbit list` command
var listCmd = &cobra.Command{
	Use:   "list [project-name]",
	Short: "List all available projects",
	Long: `Display all project names defined in the YAML configuration.

Useful for verifying your configuration is valid and seeing
available projects before running them.`,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		names, err := getProjectNamesFromConfig()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return names, cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		ui.SetQuietMode(quiet)
		ui.Header("Available Projects")

		spinner := ui.NewSpinner("Loading configuration...")
		spinner.Start()

		cfg, err := config.LoadConfig(configFile)
		spinner.Stop()

		if err != nil {
			ui.Error(fmt.Sprintf("Failed to load config: %v", err))
			return err
		}

		if len(cfg.Projects) == 0 {
			ui.Warning("No projects found in configuration.")
			fmt.Println()
			ui.Info("Get started:")
			fmt.Printf("  %s\n", ui.ColorBold("orbit init"))
			fmt.Printf("  %s %s\n", ui.ColorBold("orbit run"), ui.ColorDim("<project-name>"))
			return nil
		}

		// Sort project names for consistent output
		names := make([]string, 0, len(cfg.Projects))
		for name := range cfg.Projects {
			names = append(names, name)
		}
		sort.Strings(names)

		ui.Success(fmt.Sprintf("Found %d project(s)", len(cfg.Projects)))
		fmt.Println()

		for _, name := range names {
			proj := cfg.Projects[name]
			host := proj.SSH.Host
			if host == "" {
				host = proj.SSH.Alias
			}
			fmt.Printf("  %s %s\n", ui.ColorBold(name), ui.ColorDim(fmt.Sprintf("(→ %s@%s)", proj.SSH.User, host)))
		}

		fmt.Println()
		ui.Info("Run with:")
		fmt.Printf("  %s %s\n", ui.ColorBold("orbit run"), ui.ColorDim("<project-name>"))
		fmt.Println()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
