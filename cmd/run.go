package cmd

import (
	"fmt"

	"github.com/cristiangonsevi/orbit/internal/config"
	"github.com/cristiangonsevi/orbit/internal/executor"
	"github.com/spf13/cobra"
)

var dryRun bool

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
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectName := args[0]

		if verbose {
			fmt.Printf("Loading configuration for project %q\n", projectName)
		}

		cfg, err := config.LoadConfig(configFile)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		project, exists := cfg.Projects[projectName]
		if !exists {
			return fmt.Errorf("project %q not found in configuration", projectName)
		}

		exec := executor.New(project, verbose)

		if dryRun {
			exec.DryRun()
			return nil
		}

		if verbose {
			fmt.Printf("Starting project %q\n", projectName)
		}

		if err := exec.Run(); err != nil {
			return fmt.Errorf("project %q failed: %w", projectName, err)
		}

		fmt.Printf("Project %q completed successfully.\n", projectName)
		return nil
	},
}

func init() {
	runCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Validate the project configuration without executing changes")
	rootCmd.AddCommand(runCmd)
}
