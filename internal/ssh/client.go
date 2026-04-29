package ssh

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"

	"github.com/cristiangonsevi/orbit/internal/config"
)

// Client wraps an SSH connection and provides methods for command execution.
type Client struct {
	client *ssh.Client
	config *config.SSHConfig
}

// NewClient creates a new SSH client from the given SSH config.
// It returns the client ready for use, or an error if connection fails.
func NewClient(cfg *config.SSHConfig) (*Client, error) {
	sshCfg, err := buildSSHConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("building SSH config: %w", err)
	}

	host := cfg.Host
	if host == "" && cfg.Alias != "" {
		resolved, err := resolveSSHAlias(cfg.Alias)
		if err != nil {
			return nil, fmt.Errorf("resolving SSH alias %q: %w", cfg.Alias, err)
		}
		host = resolved
	}

	addr := net.JoinHostPort(host, "22")
	conn, err := ssh.Dial("tcp", addr, sshCfg)
	if err != nil {
		return nil, fmt.Errorf("dialing SSH %s@%s: %w", cfg.User, addr, err)
	}

	return &Client{
		client: conn,
		config: cfg,
	}, nil
}

// Close terminates the SSH connection.
func (c *Client) Close() error {
	return c.client.Close()
}

// RunCommands executes a sequence of commands in a single SSH session.
// Commands are joined with "&&" so that failure in any command stops execution.
// Returns the combined stdout+stderr output and any error.
func (c *Client) RunCommands(commands []string, verbose bool) error {
	if len(commands) == 0 {
		return nil
	}

	session, err := c.client.NewSession()
	if err != nil {
		return fmt.Errorf("creating SSH session: %w", err)
	}
	defer session.Close()

	// Set up stdout/stderr
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	// Join commands with "&&" for sequential execution with error propagation
	script := strings.Join(commands, "\n")
	if verbose {
		fmt.Fprintf(os.Stderr, "[SSH] Executing %d command(s) on %s\n", len(commands), c.config.Host)
		fmt.Fprintf(os.Stderr, "[SSH] Commands:\n%s\n", script)
	}

	if err := session.Run(script); err != nil {
		// session.Run returns *ssh.ExitError on non-zero exit
		return fmt.Errorf("remote command execution failed: %w", err)
	}

	return nil
}

// buildSSHConfig converts the project's SSHConfig into an ssh.ClientConfig.
func buildSSHConfig(cfg *config.SSHConfig) (*ssh.ClientConfig, error) {
	sshCfg := &ssh.ClientConfig{
		User:            cfg.User,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Accept any host key
		Timeout:         30 * time.Second,
	}

	switch cfg.Auth.Type {
	case "key":
		auth, err := keyAuth(cfg.Auth.KeyPath, cfg.Auth.Passphrase)
		if err != nil {
			return nil, err
		}
		sshCfg.Auth = []ssh.AuthMethod{auth}

	case "password":
		sshCfg.Auth = []ssh.AuthMethod{
			ssh.Password(cfg.Auth.Password),
		}

	default:
		return nil, fmt.Errorf("unsupported auth type: %s", cfg.Auth.Type)
	}

	return sshCfg, nil
}

// keyAuth creates an ssh.AuthMethod from a private key file path and optional passphrase.
func keyAuth(keyPath, passphrase string) (ssh.AuthMethod, error) {
	// Expand ~ to home directory
	if strings.HasPrefix(keyPath, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("getting home directory: %w", err)
		}
		keyPath = filepath.Join(home, keyPath[2:])
	}

	keyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("reading private key %q: %w", keyPath, err)
	}

	var signer ssh.Signer
	if passphrase != "" {
		signer, err = ssh.ParsePrivateKeyWithPassphrase(keyBytes, []byte(passphrase))
		if err != nil {
			return nil, fmt.Errorf("parsing encrypted private key %q: %w", keyPath, err)
		}
	} else {
		signer, err = ssh.ParsePrivateKey(keyBytes)
		if err != nil {
			// If parsing fails without a passphrase, ask interactively
			signer, err = tryInteractivePassphrase(keyBytes, keyPath)
			if err != nil {
				return nil, fmt.Errorf("parsing private key %q: %w", keyPath, err)
			}
		}
	}

	return ssh.PublicKeys(signer), nil
}

// tryInteractivePassphrase prompts the user for a passphrase if the key is encrypted.
func tryInteractivePassphrase(keyBytes []byte, keyPath string) (ssh.Signer, error) {
	fmt.Fprintf(os.Stderr, "Private key %q is encrypted.\n", keyPath)
	fmt.Fprintf(os.Stderr, "Enter passphrase for key %q: ", keyPath)
	passBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return nil, fmt.Errorf("reading passphrase: %w", err)
	}
	passphrase := string(passBytes)
	if passphrase == "" {
		return nil, fmt.Errorf("passphrase is required for encrypted key %q", keyPath)
	}
	return ssh.ParsePrivateKeyWithPassphrase(keyBytes, []byte(passphrase))
}

// resolveSSHAlias attempts to extract the HostName from ~/.ssh/config for a given alias.
// Falls back to using the alias as the hostname if parsing fails.
func resolveSSHAlias(alias string) (string, error) {
	sshConfigPath := filepath.Join(os.Getenv("HOME"), ".ssh", "config")
	if _, err := os.Stat(sshConfigPath); os.IsNotExist(err) {
		return alias, fmt.Errorf("~/.ssh/config not found, using alias %q as hostname", alias)
	}

	data, err := os.ReadFile(sshConfigPath)
	if err != nil {
		return alias, fmt.Errorf("reading ~/.ssh/config: %w, using alias %q as hostname", err, alias)
	}

	hostname := findHostNameInSSHConfig(string(data), alias)
	if hostname == "" {
		return alias, fmt.Errorf("alias %q not found in ~/.ssh/config, using it as hostname", alias)
	}

	return hostname, nil
}

// findHostNameInSSHConfig parses a simple Host → HostName mapping from SSH config text.
// This is a simplified parser; it handles the common case.
func findHostNameInSSHConfig(configText, alias string) string {
	lines := strings.Split(configText, "\n")
	inHost := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Check for Host directive
		if strings.HasPrefix(strings.ToLower(line), "host ") {
			hostPattern := strings.TrimSpace(line[5:])
			inHost = matchHostPattern(hostPattern, alias)
			continue
		}

		// If we're inside the matching Host block, look for HostName
		if inHost && strings.HasPrefix(strings.ToLower(line), "hostname ") {
			return strings.TrimSpace(line[9:])
		}
	}

	return ""
}

// matchHostPattern checks if an SSH config host pattern matches the given alias.
// Supports wildcard patterns (* and ?).
func matchHostPattern(pattern, alias string) bool {
	parts := strings.Split(pattern, " ")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if matched, _ := filepath.Match(part, alias); matched {
			return true
		}
	}
	return false
}
