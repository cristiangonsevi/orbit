package cmd

import (
	"fmt"
	"sort"

	"github.com/cristiangonsevi/orbit/internal/config"
	"github.com/spf13/cobra"
)

// listCmd represents the `orbit list` command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available projects",
	Long: `Display all project names defined in the YAML configuration.

Useful for verifying your configuration is valid and seeing
available projects before running them.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig(configFile)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		if len(cfg.Projects) == 0 {
			fmt.Println("No projects found in configuration.")
			fmt.Println("Edit the config file to define projects:")
			fmt.Printf("  %s\n", config.DefaultConfigPath())
			return nil
		}

		fmt.Println("Available projects:")
		fmt.Println()

		// Sort project names for consistent output
		names := make([]string, 0, len(cfg.Projects))
		for name := range cfg.Projects {
			names = append(names, name)
		}
		sort.Strings(names)

		for _, name := range names {
			proj := cfg.Projects[name]
			host := proj.SSH.Host
			if host == "" {
				host = proj.SSH.Alias
			}
			fmt.Printf("  %s (→ %s@%s)\n", name, proj.SSH.User, host)
		}

		fmt.Println()
		fmt.Println("Run 'orbit run <project-name>' to execute a project.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
