package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the top-level YAML configuration structure.
type Config struct {
	Projects map[string]*Project `yaml:"projects"`
}

// Project defines a complete deploy/run workflow.
type Project struct {
	SSH    *SSHConfig    `yaml:"ssh"`
	Local  *LocalConfig  `yaml:"local,omitempty"`
	Upload []UploadEntry `yaml:"upload,omitempty"`
	Remote *RemoteConfig `yaml:"remote,omitempty"`
}

// SSHConfig holds remote connection details.
type SSHConfig struct {
	Host  string     `yaml:"host,omitempty"`
	Alias string     `yaml:"alias,omitempty"`
	User  string     `yaml:"user"`
	Auth  AuthConfig `yaml:"auth"`
}

// AuthConfig holds authentication method details.
type AuthConfig struct {
	Type       string `yaml:"type"` // "key" or "password"
	KeyPath    string `yaml:"key_path,omitempty"`
	Passphrase string `yaml:"passphrase,omitempty"`
	Password   string `yaml:"password,omitempty"`
}

// LocalConfig defines local command execution settings.
type LocalConfig struct {
	WorkingDir string   `yaml:"working_dir,omitempty"`
	Before     []string `yaml:"before,omitempty"`
	After      []string `yaml:"after,omitempty"`
}

// UploadEntry represents a source → destination file transfer.
type UploadEntry struct {
	Source      string `yaml:"source"`
	Destination string `yaml:"destination"`
}

// RemoteConfig contains sequential remote commands to execute.
type RemoteConfig struct {
	Commands []string `yaml:"commands"`
}

// LoadConfig reads and parses a YAML config file from the given path.
// If path is empty, it uses the default config location.
func LoadConfig(path string) (*Config, error) {
	if path == "" {
		path = DefaultConfigPath()
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file %q: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file %q: %w", path, err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config in %q: %w", path, err)
	}

	return &cfg, nil
}

// Validate performs basic validation on the configuration.
func (c *Config) Validate() error {
	if len(c.Projects) == 0 {
		return fmt.Errorf("no projects defined in config")
	}
	for name, p := range c.Projects {
		if err := p.Validate(); err != nil {
			return fmt.Errorf("project %q: %w", name, err)
		}
	}
	return nil
}

// Validate checks that a project has the required fields set.
func (p *Project) Validate() error {
	if p.SSH == nil {
		return fmt.Errorf("ssh section is required")
	}
	if p.SSH.Host == "" && p.SSH.Alias == "" {
		return fmt.Errorf("either ssh.host or ssh.alias must be provided")
	}
	if p.SSH.Host != "" && p.SSH.Alias != "" {
		return fmt.Errorf("ssh.host and ssh.alias are mutually exclusive; use only one")
	}
	if p.SSH.User == "" {
		return fmt.Errorf("ssh.user is required")
	}
	if p.SSH.Auth.Type == "" {
		return fmt.Errorf("ssh.auth.type is required (must be 'key' or 'password')")
	}
	if p.SSH.Auth.Type != "key" && p.SSH.Auth.Type != "password" {
		return fmt.Errorf("ssh.auth.type must be 'key' or 'password', got %q", p.SSH.Auth.Type)
	}
	if p.SSH.Auth.Type == "key" && p.SSH.Alias == "" && p.SSH.Auth.KeyPath == "" {
		return fmt.Errorf("ssh.auth.key_path is required for key-based auth when no alias is used")
	}
	if p.SSH.Auth.Type == "password" && p.SSH.Auth.Password == "" {
		return fmt.Errorf("ssh.auth.password is required for password-based auth")
	}
	if p.Remote == nil || len(p.Remote.Commands) == 0 {
		return fmt.Errorf("remote.commands must have at least one command")
	}
	return nil
}

// InitConfig creates a default config file at the default path if it doesn't exist.
func InitConfig(template string) (string, error) {
	path := DefaultConfigPath()
	dir := filepath.Dir(path)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("creating config directory %q: %w", dir, err)
	}

	// Check if file already exists
	if _, err := os.Stat(path); err == nil {
		return path, fmt.Errorf("config file already exists at %q", path)
	}

	if err := os.WriteFile(path, []byte(template), 0600); err != nil {
		return "", fmt.Errorf("writing config file %q: %w", path, err)
	}

	return path, nil
}
