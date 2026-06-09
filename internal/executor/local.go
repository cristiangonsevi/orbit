package executor

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// runLocalCommands executes a series of commands locally.
// If workingDir is set, commands are executed in that directory.
// All commands are joined with " && " and run in a single shell so that
// environment variables (export) and directory changes (cd) persist across commands.
func runLocalCommands(commands []string, workingDir string, verbose bool) error {
	if len(commands) == 0 {
		return nil
	}

	// Validate working directory if specified
	if workingDir != "" {
		if err := expandAndChdir(workingDir); err != nil {
			return fmt.Errorf("changing to working directory %q: %w", workingDir, err)
		}
	}

	// Join all commands into a single shell invocation so exports and cd persist
	combined := strings.Join(commands, " && ")

	if verbose {
		fmt.Fprintf(os.Stderr, "[LOCAL] Running %d commands in single shell\n", len(commands))
		fmt.Fprintf(os.Stderr, "[LOCAL] %s\n", combined)
	}

	cmd := buildCommand(combined)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("local commands failed: %w", err)
	}

	return nil
}

// buildCommand creates an exec.Cmd from a command string.
// It uses the system shell on Unix-like systems.
func buildCommand(cmdStr string) *exec.Cmd {
	return exec.Command("sh", "-c", cmdStr)
}

// expandAndChdir expands ~ in a path and changes to that directory.
func expandAndChdir(dir string) error {
	expanded := dir
	if strings.HasPrefix(expanded, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("getting home directory: %w", err)
		}
		expanded = home + expanded[1:]
	}
	return os.Chdir(expanded)
}

// dryRunLocalCommands prints what local commands would be executed without running them.
func dryRunLocalCommands(commands []string, workingDir string) {
	if len(commands) == 0 {
		fmt.Println("[DRY-RUN] No local commands to execute")
		return
	}

	fmt.Printf("[DRY-RUN] Would execute %d local command(s)\n", len(commands))
	if workingDir != "" {
		fmt.Printf("[DRY-RUN] Working directory: %s\n", workingDir)
	}
	for i, cmd := range commands {
		fmt.Printf("  %d. %s\n", i+1, cmd)
	}
}

// restoreWorkingDir changes back to the original working directory.
// It's a no-op if target is empty.
func restoreWorkingDir(originalDir string) {
	if originalDir == "" {
		return
	}
	_ = os.Chdir(originalDir)
}
