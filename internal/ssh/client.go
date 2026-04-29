package ssh

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
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
	// If an alias is used without explicit host, resolve host and IdentityFile
	// from ~/.ssh/config
	var resolvedKeyPath string
	host := cfg.Host
	if host == "" && cfg.Alias != "" {
		aliasInfo, err := resolveSSHAlias(cfg.Alias)
		if err != nil {
			return nil, fmt.Errorf("resolving SSH alias %q: %w", cfg.Alias, err)
		}
		host = aliasInfo.hostname
		resolvedKeyPath = aliasInfo.identityFile
	}

	// If no explicit key_path but we resolved one from the SSH config, use it
	sshAuthCfg := *cfg
	if cfg.Auth.KeyPath == "" && resolvedKeyPath != "" {
		sshAuthCfg.Auth.KeyPath = resolvedKeyPath
	}

	sshCfg, err := buildSSHConfig(&sshAuthCfg)
	if err != nil {
		return nil, fmt.Errorf("building SSH config: %w", err)
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
		auths, err := buildKeyAuths(cfg.Auth.KeyPath, cfg.Auth.Passphrase)
		if err != nil {
			return nil, err
		}
		sshCfg.Auth = auths

	case "password":
		sshCfg.Auth = []ssh.AuthMethod{
			ssh.Password(cfg.Auth.Password),
		}

	default:
		return nil, fmt.Errorf("unsupported auth type: %s", cfg.Auth.Type)
	}

	return sshCfg, nil
}

// buildKeyAuths returns one or more AuthMethods for key-based authentication.
// If keyPath is empty, it tries the SSH agent first, then falls back to
// common default key paths.
func buildKeyAuths(keyPath, passphrase string) ([]ssh.AuthMethod, error) {
	var auths []ssh.AuthMethod

	// Always try the SSH agent first (if available) — this is the standard
	// behavior when using `ssh <alias>` from the command line.
	agentAuth := sshAgentAuth()
	if agentAuth != nil {
		auths = append(auths, agentAuth)
	}

	if keyPath != "" {
		auth, err := keyAuth(keyPath, passphrase)
		if err != nil {
			return nil, err
		}
		auths = append(auths, auth)
	} else {
		// When no key_path is specified, try common default key locations
		defaultPaths := []string{
			filepath.Join(os.Getenv("HOME"), ".ssh", "id_rsa"),
			filepath.Join(os.Getenv("HOME"), ".ssh", "id_ed25519"),
			filepath.Join(os.Getenv("HOME"), ".ssh", "id_ecdsa"),
		}
		for _, p := range defaultPaths {
			if _, err := os.Stat(p); err == nil {
				auth, err := keyAuth(p, passphrase)
				if err != nil {
					// If the key is encrypted and passphrase is wrong, skip it
					// rather than failing — the agent might still work
					if strings.Contains(err.Error(), "passphrase") {
						continue
					}
					return nil, err
				}
				auths = append(auths, auth)
				break
			}
		}
	}

	if len(auths) == 0 {
		return nil, fmt.Errorf("no SSH authentication method available: provide a key_path or verify your SSH agent is running")
	}

	return auths, nil
}

// sshAgentAuth attempts to connect to the local SSH agent and returns an
// AuthMethod that delegates to it. Returns nil if the agent is unavailable.
func sshAgentAuth() ssh.AuthMethod {
	sshAgentConn, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
	if err != nil {
		return nil
	}
	sshAgent := agent.NewClient(sshAgentConn)
	return ssh.PublicKeysCallback(sshAgent.Signers)
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

// aliasInfo holds resolved SSH configuration for a given alias.
type aliasInfo struct {
	hostname     string
	identityFile string
}

// resolveSSHAlias attempts to extract the HostName and IdentityFile from
// ~/.ssh/config for a given alias.
// Falls back to using the alias as the hostname if parsing fails.
func resolveSSHAlias(alias string) (*aliasInfo, error) {
	sshConfigPath := filepath.Join(os.Getenv("HOME"), ".ssh", "config")
	info := &aliasInfo{hostname: alias}

	if _, err := os.Stat(sshConfigPath); os.IsNotExist(err) {
		return info, fmt.Errorf("~/.ssh/config not found, using alias %q as hostname", alias)
	}

	data, err := os.ReadFile(sshConfigPath)
	if err != nil {
		return info, fmt.Errorf("reading ~/.ssh/config: %w, using alias %q as hostname", err, alias)
	}

	parsed := parseSSHConfigBlock(string(data), alias)
	if parsed != nil {
		if parsed.hostname != "" {
			info.hostname = parsed.hostname
		}
		if parsed.identityFile != "" {
			info.identityFile = parsed.identityFile
		}
	}

	return info, nil
}

// parseSSHConfigBlock parses a ~/.ssh/config block for the given alias
// and returns hostname and identityFile if found.
func parseSSHConfigBlock(configText, alias string) *aliasInfo {
	lines := strings.Split(configText, "\n")
	inHost := false
	result := &aliasInfo{}

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Check for Host directive
		lower := strings.ToLower(line)
		if strings.HasPrefix(lower, "host ") {
			hostPattern := strings.TrimSpace(line[5:])
			inHost = matchHostPattern(hostPattern, alias)
			continue
		}

		if !inHost {
			continue
		}

		// Inside the matching Host block, look for HostName and IdentityFile
		if strings.HasPrefix(lower, "hostname ") {
			result.hostname = strings.TrimSpace(line[9:])
		} else if strings.HasPrefix(lower, "identityfile ") {
			result.identityFile = strings.TrimSpace(line[13:])
		}
	}

	return result
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
