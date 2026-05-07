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
	// quiet suppresses animations and progress indicators
	quiet bool
	// Version is set at build time via ldflags
	Version = "0.0.4"
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
	Version:                Version,
	SilenceErrors:          true,
	SilenceUsage:           true,
	BashCompletionFunction: bashCompletionFunction,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return []string{"init", "list", "run", "validate", "completion", "version"}, cobra.ShellCompDirectiveNoFileComp
	},
}

const (
	ansiRed   = "\033[31m"
	ansiReset = "\033[0m"
)

// bashCompletionFunction handles bash completion for orbit commands
// It provides dynamic completion for project names based on the config file
const bashCompletionFunction = `
# Orbit bash completion function
_orbit_completion() {
    local cur prev words cword
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    # Main command completion
    if [[ $cword -eq 1 ]]; then
        COMPREPLY=($(compgen -W "init list run validate completion version" -- "$cur"))
        return 0
    fi

    # Subcommand completion
    case "${prev}" in
        run)
            # Get project names from config file
            local config_path="${HOME}/.config/orbit/config.yaml"
            if [[ -f "$config_path" ]]; then
                local projects=$(grep -E '^\s*[a-zA-Z0-9_-]+:' "$config_path" 2>/dev/null | sed 's/:.*//' | tr -d ' ')
                COMPREPLY=($(compgen -W "$projects" -- "$cur"))
            fi
            return 0
            ;;
        completion)
            COMPREPLY=($(compgen -W "bash zsh fish powershell" -- "$cur"))
            return 0
            ;;
        --config|-v|--verbose|--quiet)
            _filedir
            return 0
            ;;
    esac

    return 0
}

# Register the completion function
complete -F _orbit_completion orbit
`

func printError(err error) {
	fmt.Fprintln(os.Stderr, ansiRed+err.Error()+ansiReset)
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		printError(err)
		os.Exit(1)
	}
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "Path to config file (default: ~/.config/orbit/config.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output for debugging")
	rootCmd.PersistentFlags().BoolVar(&quiet, "quiet", false, "Suppress animations and progress indicators (useful for CI/CD)")
}

// IsQuiet returns whether quiet mode is enabled
func IsQuiet() bool {
	return quiet
}
