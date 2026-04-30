package config

import (
	"os"
	"path/filepath"
)

const (
	// DefaultConfigDirName is the config directory name under the user's config home.
	DefaultConfigDirName = ".orbit"
	// DefaultConfigFileName is the default YAML config file name.
	DefaultConfigFileName = "config.yaml"
	// DefaultConfigDirEnv is the env var for the XDG config home directory.
	DefaultConfigDirEnv = "XDG_CONFIG_HOME"
)

// DefaultConfigDir returns the directory where the config file is stored.
// It respects the XDG_CONFIG_HOME environment variable, falling back to ~/.config.
func DefaultConfigDir() string {
	if dir := os.Getenv(DefaultConfigDirEnv); dir != "" {
		return filepath.Join(dir, DefaultConfigDirName)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".config", DefaultConfigDirName)
	}
	return filepath.Join(home, ".config", DefaultConfigDirName)
}

// DefaultConfigPath returns the full path to the default config file.
func DefaultConfigPath() string {
	return filepath.Join(DefaultConfigDir(), DefaultConfigFileName)
}
