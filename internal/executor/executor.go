package executor

import (
	"fmt"
	"os"
	"strings"

	"github.com/cristiangonsevi/orbit/internal/config"
	"github.com/cristiangonsevi/orbit/internal/ssh"
	"github.com/cristiangonsevi/orbit/internal/uploader"
)

// Executor coordinates the full project workflow:
// local before → ssh upload → remote commands → local after.
type Executor struct {
	project *config.Project
	verbose bool
	// savedWorkingDir stores the original directory to restore after execution
	savedWorkingDir string
}

// New creates a new Executor for the given project.
func New(project *config.Project, verbose bool) *Executor {
	return &Executor{
		project: project,
		verbose: verbose,
	}
}

// Run executes the full project workflow and returns any error.
func (e *Executor) Run() error {
	if e.verbose {
		fmt.Fprintf(os.Stderr, "[EXEC] Starting workflow for project\n")
	}

	// Save current working directory to restore later
	wd, err := os.Getwd()
	if err == nil {
		e.savedWorkingDir = wd
	}

	// Step 1: Execute local `before` commands
	if err := e.runBefore(); err != nil {
		return fmt.Errorf("before commands: %w", err)
	}

	// Step 2: Connect to SSH
	sshClient, err := ssh.NewClient(e.project.SSH)
	if err != nil {
		return fmt.Errorf("SSH connection: %w", err)
	}
	defer func() {
		if closeErr := sshClient.Close(); closeErr != nil && e.verbose {
			fmt.Fprintf(os.Stderr, "[EXEC] Warning: error closing SSH connection: %v\n", closeErr)
		}
	}()

	// Step 3: Upload files (if any)
	if err := e.runUpload(sshClient); err != nil {
		return fmt.Errorf("upload: %w", err)
	}

	// Step 4: Execute remote commands
	if err := e.runRemote(sshClient); err != nil {
		return fmt.Errorf("remote commands: %w", err)
	}

	// Restore working directory before running local after commands
	e.restoreWD()

	// Step 5: Execute local `after` commands
	if err := e.runAfter(); err != nil {
		return fmt.Errorf("after commands: %w", err)
	}

	if e.verbose {
		fmt.Fprintf(os.Stderr, "[EXEC] Workflow completed successfully\n")
	}

	return nil
}

// DryRun prints what would be executed without actually running anything.
func (e *Executor) DryRun() {
	fmt.Println("[DRY-RUN] Project workflow plan:")
	fmt.Println(strings.Repeat("=", 60))

	// Local before
	fmt.Println("\n📋 Step 1: Local before commands")
	dryRunLocalCommands(e.project.Local.Before, e.project.Local.WorkingDir)

	// SSH connection
	fmt.Println("\n📋 Step 2: SSH connection")
	fmt.Printf("  Host: %s\n", e.project.SSH.Host)
	if e.project.SSH.Alias != "" {
		fmt.Printf("  Alias: %s\n", e.project.SSH.Alias)
	}
	fmt.Printf("  User: %s\n", e.project.SSH.User)
	fmt.Printf("  Auth type: %s\n", e.project.SSH.Auth.Type)

	// Upload
	uploader.DryRunUploads(e.project.Upload)

	// Remote commands
	fmt.Println("\n📋 Step 4: Remote commands")
	if e.project.Remote != nil && len(e.project.Remote.Commands) > 0 {
		for i, cmd := range e.project.Remote.Commands {
			fmt.Printf("  %d. %s\n", i+1, cmd)
		}
	} else {
		fmt.Println("  (none)")
	}

	// Local after
	fmt.Println("\n📋 Step 5: Local after commands")
	dryRunLocalCommands(e.project.Local.After, e.project.Local.WorkingDir)

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("[DRY-RUN] No changes were made. Run without --dry-run to execute.")
}

// runBefore executes the local `before` commands.
func (e *Executor) runBefore() error {
	if e.project.Local == nil || len(e.project.Local.Before) == 0 {
		if e.verbose {
			fmt.Fprintf(os.Stderr, "[EXEC] No before commands to execute\n")
		}
		return nil
	}

	if e.verbose {
		fmt.Fprintf(os.Stderr, "[EXEC] Running %d before command(s)\n", len(e.project.Local.Before))
	}

	return runLocalCommands(e.project.Local.Before, e.project.Local.WorkingDir, e.verbose)
}

// runUpload uploads files to the remote server.
func (e *Executor) runUpload(client *ssh.Client) error {
	if e.project.Upload == nil || len(e.project.Upload) == 0 {
		if e.verbose {
			fmt.Fprintf(os.Stderr, "[EXEC] No uploads configured, skipping\n")
		}
		return nil
	}

	if e.verbose {
		fmt.Fprintf(os.Stderr, "[EXEC] Uploading %d file(s)/dir(s)\n", len(e.project.Upload))
	}

	return uploader.UploadFiles(client, e.project.Upload, e.verbose)
}

// runRemote executes commands on the remote server.
func (e *Executor) runRemote(client *ssh.Client) error {
	if e.project.Remote == nil || len(e.project.Remote.Commands) == 0 {
		return fmt.Errorf("no remote commands configured")
	}

	if e.verbose {
		fmt.Fprintf(os.Stderr, "[EXEC] Executing %d remote command(s)\n", len(e.project.Remote.Commands))
	}

	return client.RunCommands(e.project.Remote.Commands, e.verbose)
}

// runAfter executes the local `after` commands.
func (e *Executor) runAfter() error {
	if e.project.Local == nil || len(e.project.Local.After) == 0 {
		if e.verbose {
			fmt.Fprintf(os.Stderr, "[EXEC] No after commands to execute\n")
		}
		return nil
	}

	if e.verbose {
		fmt.Fprintf(os.Stderr, "[EXEC] Running %d after command(s)\n", len(e.project.Local.After))
	}

	return runLocalCommands(e.project.Local.After, e.project.Local.WorkingDir, e.verbose)
}

// restoreWD restores the original working directory.
func (e *Executor) restoreWD() {
	restoreWorkingDir(e.savedWorkingDir)
}
