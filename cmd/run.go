package cmd

import (
	"fmt"

	"github.com/cristiangonsevi/orbit/internal/config"
	"github.com/cristiangonsevi/orbit/internal/executor"
	"github.com/cristiangonsevi/orbit/internal/ui"
	"github.com/spf13/cobra"
)

// showWarnings displays configuration warnings for a project.
func showWarnings(cfg *config.Config, projectName string) {
	warnings := cfg.ValidateDeep()
	if len(warnings) == 0 {
		return
	}
	for _, w := range warnings {
		if w.Project == projectName {
			ui.Warning(fmt.Sprintf("Project %q: %s", w.Project, w.Message))
		}
	}
}

var dryRun bool

// getProjectNames returns all project names from the config file
func getProjectNames() ([]string, error) {
	return getProjectNamesFromConfig()
}

// runCmd represents the `orbit run` command
var runCmd = &cobra.Command{
	Use:   "run <project-name>",
	Short: "Execute a project workflow",
	Long: `Run a deployment or automation workflow for a named project.

The workflow consists of:
  1. Local "before" commands
  2. SSH connection to the remote server
  3. File uploads (if configured)
  4. Remote commands execution
  5. Local "after" commands

Use the --dry-run flag to see what would be executed without
actually running any commands.`,
	Args:         cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		names, err := getProjectNames()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return names, cobra.ShellCompDirectiveNoFileComp
	},
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		ui.SetQuietMode(quiet)
		projectName := args[0]

		cfg, err := config.LoadConfig(configFile)
		if err != nil {
			ui.Error(fmt.Sprintf("Failed to load config: %v", err))
			return err
		}

		showWarnings(cfg, projectName)

		project, exists := cfg.Projects[projectName]
		if !exists {
			ui.Error(fmt.Sprintf("Project %q not found in configuration", projectName))
			ui.Info("Run 'orbit list' to see available projects")
			return fmt.Errorf("project %q not found", projectName)
		}

		ui.Header(fmt.Sprintf("Running: %s", projectName))
		exec := executor.New(project, verbose)

		if dryRun {
			ui.SubHeader("Dry Run Mode")
			ui.Info(fmt.Sprintf("Showing workflow for project: %s", projectName))
			fmt.Println()
			exec.DryRun()
			return nil
		}

		// Run the project with animated feedback
		if err := exec.Run(); err != nil {
			ui.Error(fmt.Sprintf("Project %q failed", projectName))
			return err
		}

		fmt.Println()
		ui.Separator()
		ui.Success(fmt.Sprintf("Project %q completed successfully!", projectName))
		ui.Separator()
		fmt.Println()

		return nil
	},
}

func init() {
	runCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Validate the project configuration without executing changes")
	rootCmd.AddCommand(runCmd)
}
