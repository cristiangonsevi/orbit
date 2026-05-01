package executor

import (
	"testing"

	"github.com/cristiangonsevi/orbit/internal/config"
)

func TestExecutorNew(t *testing.T) {
	project := &config.Project{
		SSH: &config.SSHConfig{
			Host: "example.com",
			User: "deployer",
			Auth: config.AuthConfig{
				Type: "key",
			},
		},
		Remote: &config.RemoteConfig{
			Commands: []string{"ls"},
		},
	}

	exec := New(project, true)
	if exec == nil {
		t.Error("New returned nil")
	}
	if exec.project != project {
		t.Error("Project not set correctly")
	}
	if !exec.verbose {
		t.Error("Verbose not set correctly")
	}
}

func TestDryRunLocalCommands(t *testing.T) {
	t.Run("empty commands", func(t *testing.T) {
		dryRunLocalCommands([]string{}, "")
	})

	t.Run("with commands", func(t *testing.T) {
		commands := []string{"npm ci", "npm run build"}
		workingDir := "./app"
		dryRunLocalCommands(commands, workingDir)
	})
}

func TestRestoreWorkingDir(t *testing.T) {
	t.Run("empty dir", func(t *testing.T) {
		restoreWorkingDir("")
	})

	t.Run("restore to specific dir", func(t *testing.T) {
		restoreWorkingDir("/nonexistent/path")
	})
}

func TestBuildCommand(t *testing.T) {
	cmd := buildCommand("echo hello")
	if cmd == nil {
		t.Error("buildCommand returned nil")
	}
	// The command uses sh -c, so Args[0] is "sh", Args[1] is "-c"
	if len(cmd.Args) < 2 {
		t.Error("Expected at least 2 args (sh, -c)")
	}
	if cmd.Args[0] != "sh" || cmd.Args[1] != "-c" {
		t.Errorf("Expected 'sh -c', got %v", cmd.Args[:2])
	}
}

func TestExpandAndChdir(t *testing.T) {
	t.Run("empty path returns error", func(t *testing.T) {
		// Empty path should return an error
		err := expandAndChdir("")
		if err == nil {
			t.Error("expandAndChdir('') expected error, got nil")
		}
	})

	t.Run("expand tilde", func(t *testing.T) {
		// Creating and changing to a temp dir
		tmpDir := t.TempDir()
		err := expandAndChdir(tmpDir)
		if err != nil {
			t.Errorf("expandAndChdir(%s) error: %v", tmpDir, err)
		}
	})
}
