package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config with all fields",
			config: Config{
				Projects: map[string]*Project{
					"test-project": {
						SSH: &SSHConfig{
							Host: "example.com",
							User: "deployer",
							Auth: AuthConfig{
								Type:    "key",
								KeyPath: "~/.ssh/id_rsa",
							},
						},
						Remote: &RemoteConfig{
							Commands: []string{"ls -la"},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "empty projects",
			config: Config{
				Projects: map[string]*Project{},
			},
			wantErr: true,
		},
		{
			name: "nil ssh",
			config: Config{
				Projects: map[string]*Project{
					"test-project": {
						SSH:    nil,
						Remote: &RemoteConfig{Commands: []string{"ls"}},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "missing host and alias",
			config: Config{
				Projects: map[string]*Project{
					"test-project": {
						SSH: &SSHConfig{
							User: "deployer",
							Auth: AuthConfig{Type: "key"},
						},
						Remote: &RemoteConfig{Commands: []string{"ls"}},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "both host and alias set",
			config: Config{
				Projects: map[string]*Project{
					"test-project": {
						SSH: &SSHConfig{
							Host:  "example.com",
							Alias: "myalias",
							User:  "deployer",
							Auth:  AuthConfig{Type: "key"},
						},
						Remote: &RemoteConfig{Commands: []string{"ls"}},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "missing user",
			config: Config{
				Projects: map[string]*Project{
					"test-project": {
						SSH: &SSHConfig{
							Host: "example.com",
							Auth: AuthConfig{Type: "key"},
						},
						Remote: &RemoteConfig{Commands: []string{"ls"}},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "missing auth type",
			config: Config{
				Projects: map[string]*Project{
					"test-project": {
						SSH: &SSHConfig{
							Host: "example.com",
							User: "deployer",
							Auth: AuthConfig{Type: ""},
						},
						Remote: &RemoteConfig{Commands: []string{"ls"}},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid auth type",
			config: Config{
				Projects: map[string]*Project{
					"test-project": {
						SSH: &SSHConfig{
							Host: "example.com",
							User: "deployer",
							Auth: AuthConfig{Type: "oauth"},
						},
						Remote: &RemoteConfig{Commands: []string{"ls"}},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "key auth without key_path or alias",
			config: Config{
				Projects: map[string]*Project{
					"test-project": {
						SSH: &SSHConfig{
							Host: "example.com",
							User: "deployer",
							Auth: AuthConfig{Type: "key", KeyPath: ""},
						},
						Remote: &RemoteConfig{Commands: []string{"ls"}},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "password auth without password",
			config: Config{
				Projects: map[string]*Project{
					"test-project": {
						SSH: &SSHConfig{
							Host: "example.com",
							User: "deployer",
							Auth: AuthConfig{Type: "password"},
						},
						Remote: &RemoteConfig{Commands: []string{"ls"}},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "missing remote commands",
			config: Config{
				Projects: map[string]*Project{
					"test-project": {
						SSH: &SSHConfig{
							Host: "example.com",
							User: "deployer",
							Auth: AuthConfig{Type: "key"},
						},
						Remote: &RemoteConfig{Commands: []string{}},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "using SSH alias",
			config: Config{
				Projects: map[string]*Project{
					"test-project": {
						SSH: &SSHConfig{
							Alias: "myalias",
							User:  "deployer",
							Auth:  AuthConfig{Type: "key"},
						},
						Remote: &RemoteConfig{Commands: []string{"ls"}},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	validConfig := `projects:
  test-project:
    ssh:
      host: example.com
      user: deployer
      auth:
        type: key
        key_path: ~/.ssh/id_rsa
    remote:
      commands:
        - ls -la
`
	err := os.WriteFile(configPath, []byte(validConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp config: %v", err)
	}

	t.Run("load valid config", func(t *testing.T) {
		cfg, err := LoadConfig(configPath)
		if err != nil {
			t.Errorf("LoadConfig() error = %v", err)
			return
		}
		if len(cfg.Projects) != 1 {
			t.Errorf("Expected 1 project, got %d", len(cfg.Projects))
		}
	})

	t.Run("file not found", func(t *testing.T) {
		_, err := LoadConfig("/nonexistent/path.yaml")
		if err == nil {
			t.Error("Expected error for nonexistent file")
		}
	})

	t.Run("invalid yaml", func(t *testing.T) {
		invalidPath := filepath.Join(tmpDir, "invalid.yaml")
		os.WriteFile(invalidPath, []byte("invalid: yaml: content: ["), 0644)
		_, err := LoadConfig(invalidPath)
		if err == nil {
			t.Error("Expected error for invalid YAML")
		}
	})
}

func TestDefaultConfigDir(t *testing.T) {
	t.Run("with XDG_CONFIG_HOME", func(t *testing.T) {
		os.Setenv("XDG_CONFIG_HOME", "/custom/config")
		defer os.Unsetenv("XDG_CONFIG_HOME")

		dir := DefaultConfigDir()
		if dir != "/custom/config/.orbit" {
			t.Errorf("Expected /custom/config/.orbit, got %s", dir)
		}
	})

	t.Run("without XDG_CONFIG_HOME", func(t *testing.T) {
		os.Unsetenv("XDG_CONFIG_HOME")
		dir := DefaultConfigDir()
		if dir == "" {
			t.Error("Expected non-empty default config dir")
		}
	})
}

func TestProjectValidation(t *testing.T) {
	tests := []struct {
		name    string
		project Project
		wantErr bool
	}{
		{
			name: "minimal valid project with alias",
			project: Project{
				SSH: &SSHConfig{
					Alias: "myalias",
					User:  "user",
					Auth:  AuthConfig{Type: "key"},
				},
				Remote: &RemoteConfig{Commands: []string{"echo hello"}},
			},
			wantErr: false,
		},
		{
			name: "empty remote commands",
			project: Project{
				SSH: &SSHConfig{
					Host: "example.com",
					User: "user",
					Auth: AuthConfig{Type: "key"},
				},
				Remote: &RemoteConfig{Commands: []string{}},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.project.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Project.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
