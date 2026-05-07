package executor

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cristiangonsevi/orbit/internal/config"
	"github.com/cristiangonsevi/orbit/internal/logging"
	"github.com/cristiangonsevi/orbit/internal/ssh"
	"github.com/cristiangonsevi/orbit/internal/ui"
	"github.com/cristiangonsevi/orbit/internal/uploader"
)

// Executor coordinates the full project workflow:
// local before → ssh upload → remote commands → local after.
type Executor struct {
	project         *config.Project
	projectName     string
	verbose         bool
	savedWorkingDir string
	sshClient       *ssh.Client
	logger          *logging.Logger
	startTime       time.Time
}

// New creates a new Executor for the given project.
func New(project *config.Project, projectName string, verbose bool) *Executor {
	return &Executor{
		project:     project,
		projectName: projectName,
		verbose:     verbose,
	}
}

// Run executes the full project workflow and returns any error.
func (e *Executor) Run() error {
	wd, err := os.Getwd()
	if err == nil {
		e.savedWorkingDir = wd
	}

	// Create logger
	e.logger, err = logging.New(e.projectName, e.sshHost(), e.project.Remote.Commands)
	if err != nil {
		ui.Warning(fmt.Sprintf("Failed to create logger: %v", err))
	}
	e.startTime = time.Now()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	doneCh := make(chan struct{}, 1)

	type execResult struct {
		err error
	}
	resultCh := make(chan execResult, 1)

	go func() {
		resultCh <- execResult{err: e.execWorkflow()}
		close(doneCh)
	}()

	select {
	case res := <-resultCh:
		signal.Stop(sigCh)
		e.logResult(res.err)
		return res.err
	case sig := <-sigCh:
		signal.Stop(sigCh)
		ui.Error(fmt.Sprintf("Interrupted by %s", sig))
		e.cleanup()
		e.logResult(fmt.Errorf("interrupted by %s", sig))
		return fmt.Errorf("interrupted by %s", sig)
	}
}

// execWorkflow executes the actual workflow steps.
func (e *Executor) execWorkflow() error {
	// Step 1: Execute local `before` commands
	ui.Step(1, 5, "Running local before commands")
	spinner := ui.NewSpinner("Executing...")
	spinner.Start()

	if err := e.runBefore(); err != nil {
		spinner.StopWithError("Failed")
		return fmt.Errorf("before commands: %w", err)
	}
	spinner.StopWithSuccess("Done")
	ui.Separator()

	// Step 2: Connect to SSH
	ui.Step(2, 5, "Connecting via SSH")
	spinner = ui.NewSpinner(fmt.Sprintf("Connecting to %s...", e.project.SSH.Host))
	spinner.Start()

	sshClient, err := ssh.NewClient(e.project.SSH)
	if err != nil {
		spinner.StopWithError("Connection failed")
		return fmt.Errorf("SSH connection: %w", err)
	}
	e.sshClient = sshClient
	spinner.StopWithSuccess("Connected")
	ui.Separator()

	// Step 3: Upload files (if any)
	if e.project.Upload != nil && len(e.project.Upload) > 0 {
		ui.Step(3, 5, "Uploading files")
		spinner = ui.NewSpinner(fmt.Sprintf("Transferring %d file(s)...", len(e.project.Upload)))
		spinner.Start()

		if err := e.runUpload(sshClient); err != nil {
			spinner.StopWithError("Upload failed")
			return fmt.Errorf("upload: %w", err)
		}
		spinner.StopWithSuccess("Upload complete")
		ui.Separator()
	} else {
		ui.Step(3, 5, "Skipping uploads (none configured)")
		ui.Info("No files to upload")
		ui.Separator()
	}

	// Step 4: Execute remote commands
	ui.Step(4, 5, "Executing remote commands")
	spinner = ui.NewSpinner("Running on remote server...")
	spinner.Start()

	if err := e.runRemote(sshClient); err != nil {
		spinner.StopWithError("Remote execution failed")
		return fmt.Errorf("remote commands: %w", err)
	}
	spinner.StopWithSuccess("Remote commands completed")
	ui.Separator()

	// Step 5: Execute local `after` commands
	e.restoreWD()
	ui.Step(5, 5, "Running local after commands")
	spinner = ui.NewSpinner("Executing...")
	spinner.Start()

	if err := e.runAfter(); err != nil {
		spinner.StopWithError("Failed")
		return fmt.Errorf("after commands: %w", err)
	}
	spinner.StopWithSuccess("Done")

	return nil
}

// cleanup performs a clean shutdown of the executor.
func (e *Executor) cleanup() {
	if e.sshClient != nil {
		if err := e.sshClient.Close(); err != nil && e.verbose {
			fmt.Fprintf(os.Stderr, "[EXEC] Warning: error closing SSH connection: %v\n", err)
		}
		e.sshClient = nil
	}

	e.restoreWD()
}

// DryRun prints what would be executed without actually running anything.
func (e *Executor) DryRun() {
	ui.Header("Dry Run: Workflow Plan")

	// Local before
	ui.SubHeader("Step 1: Local before commands")
	if e.project.Local == nil || len(e.project.Local.Before) == 0 {
		ui.Info("No before commands")
	} else {
		if e.project.Local.WorkingDir != "" {
			fmt.Printf("  Working dir: %s\n", ui.ColorCyan(e.project.Local.WorkingDir))
		}
		for i, cmd := range e.project.Local.Before {
			fmt.Printf("  %d. %s\n", i+1, ui.ColorDim(cmd))
		}
	}
	fmt.Println()

	// SSH connection
	ui.SubHeader("Step 2: SSH connection")
	host := e.project.SSH.Host
	if host == "" {
		host = e.project.SSH.Alias
	}
	fmt.Printf("  Host: %s\n", ui.ColorCyan(host))
	fmt.Printf("  User: %s\n", ui.ColorCyan(e.project.SSH.User))
	fmt.Printf("  Auth: %s\n", ui.ColorYellow(e.project.SSH.Auth.Type))
	fmt.Println()

	// Upload
	ui.SubHeader("Step 3: File uploads")
	if len(e.project.Upload) == 0 {
		ui.Info("No uploads configured")
	} else {
		for i, entry := range e.project.Upload {
			fmt.Printf("  %d. %s %s %s\n", i+1, ui.ColorDim(entry.Source), ui.ColorCyan("→"), ui.ColorDim(entry.Destination))
		}
	}
	fmt.Println()

	// Remote commands
	ui.SubHeader("Step 4: Remote commands")
	if e.project.Remote != nil && len(e.project.Remote.Commands) > 0 {
		for i, cmd := range e.project.Remote.Commands {
			fmt.Printf("  %d. %s\n", i+1, ui.ColorDim(cmd))
		}
	} else {
		ui.Warning("No remote commands configured")
	}
	fmt.Println()

	// Local after
	ui.SubHeader("Step 5: Local after commands")
	if e.project.Local == nil || len(e.project.Local.After) == 0 {
		ui.Info("No after commands")
	} else {
		for i, cmd := range e.project.Local.After {
			fmt.Printf("  %d. %s\n", i+1, ui.ColorDim(cmd))
		}
	}
	fmt.Println()

	ui.Separator()
	ui.Info("Dry run complete. Run without --dry-run to execute.")
	ui.Separator()
	fmt.Println()
}

// runBefore executes the local `before` commands.
func (e *Executor) runBefore() error {
	if e.project.Local == nil || len(e.project.Local.Before) == 0 {
		return nil
	}

	return runLocalCommands(e.project.Local.Before, e.project.Local.WorkingDir, e.verbose)
}

// runUpload uploads files to the remote server.
func (e *Executor) runUpload(client *ssh.Client) error {
	if e.project.Upload == nil || len(e.project.Upload) == 0 {
		return nil
	}

	return uploader.UploadFiles(client, e.project.Upload, e.verbose)
}

// runRemote executes commands on the remote server.
func (e *Executor) runRemote(client *ssh.Client) error {
	if e.project.Remote == nil || len(e.project.Remote.Commands) == 0 {
		return fmt.Errorf("no remote commands configured")
	}

	return client.RunCommands(e.project.Remote.Commands, e.verbose)
}

// runAfter executes the local `after` commands.
func (e *Executor) runAfter() error {
	if e.project.Local == nil || len(e.project.Local.After) == 0 {
		return nil
	}

	return runLocalCommands(e.project.Local.After, e.project.Local.WorkingDir, e.verbose)
}

// restoreWD restores the original working directory.
func (e *Executor) restoreWD() {
	restoreWorkingDir(e.savedWorkingDir)
}

// sshHost returns the SSH host for logging.
func (e *Executor) sshHost() string {
	host := e.project.SSH.Host
	if host == "" {
		host = e.project.SSH.Alias
	}
	return fmt.Sprintf("%s@%s", e.project.SSH.User, host)
}

// logResult writes the execution result to the log file.
func (e *Executor) logResult(runErr error) {
	if e.logger == nil {
		return
	}
	duration := time.Since(e.startTime)
	entry := logging.Entry{
		Timestamp:   e.startTime.Format(time.RFC3339),
		ProjectName: e.projectName,
		UserHost:    e.sshHost(),
		Commands:    e.project.Remote.Commands,
		Success:    runErr == nil,
		Duration:   duration.String(),
	}
	if runErr != nil {
		entry.Error = runErr.Error()
	}
	if err := e.logger.Log(entry); err != nil {
		fmt.Fprintf(os.Stderr, "[LOG] Warning: failed to write log: %v\n", err)
	}
	if err := e.logger.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "[LOG] Warning: failed to close log: %v\n", err)
	}
}
