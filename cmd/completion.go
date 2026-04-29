package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// completionCmd represents the `orbit completion` command
var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Long: `Generate shell autocompletion scripts for orbit commands.

To use the completions, source the output in your shell profile:

  Bash:
    orbit completion bash > /etc/bash_completion.d/orbit
    source ~/.bashrc

  Zsh:
    orbit completion zsh > ~/.oh-my-zsh/completions/_orbit
    source ~/.zshrc

  Fish:
    orbit completion fish > ~/.config/fish/completions/orbit.fish
    source ~/.config/fish/config.fish

  PowerShell:
    orbit completion powershell > orbit.ps1
    . .\orbit.ps1`,
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
	RunE: func(cmd *cobra.Command, args []string) error {
		shell := args[0]

		switch shell {
		case "bash":
			return rootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			return rootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			return rootCmd.GenFishCompletion(os.Stdout, true)
		case "powershell":
			return rootCmd.GenPowerShellCompletion(os.Stdout)
		default:
			return fmt.Errorf("unsupported shell: %s (supported: bash, zsh, fish, powershell)", shell)
		}
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
