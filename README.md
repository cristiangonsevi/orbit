# Orbit

A CLI tool for automating remote server workflows. Connect via SSH, upload files, and run commands — all from configurable YAML files.

Built with Go and Cobra.

## Features

- Connect to remote servers via SSH (password or key)
- Use SSH aliases from your `~/.ssh/config`
- Upload files and directories to remote servers
- Run local commands before and after remote execution
- Define multiple projects in a single YAML config

## Quick Start

```bash
# Install
curl -fsSL https://raw.githubusercontent.com/cristiangonsevi/orbit/refs/heads/master/scripts/install.sh | bash

# Or build from source
go install github.com/cristiangonsevi/orbit@latest

# Initialize config
orbit init

# Run a project
orbit run my-project
```

## Features

### Upload Progress
During file transfers, Orbit displays real-time progress:
```
Uploading myfile.tar.gz  45.2% [████████░░░░] 123.5/275.0 MiB  4.2 MiB/s
```

### SSH Retry
If an SSH connection fails, Orbit automatically retries with exponential backoff (3 attempts: 1s, 2s, 4s delays).

### Persistent Logging
Every execution is logged to `~/.local/share/orbit/logs/` as JSON files:
```json
{"timestamp":"2026-05-07T14:32:01Z","project_name":"myproject","user_host":"admin@example.com","commands":["deploy.sh"],"success":true,"duration":"45.2s"}
```
Logs are rotated automatically, keeping the last 30 files.

## Installation

Download a binary for your platform from the releases page, or use the install script:

```bash
curl -fsSL https://raw.githubusercontent.com/cristiangonsevi/orbit/refs/heads/master/scripts/install.sh | bash
```

The binary installs to `~/.local/bin/orbit` by default.

## Updating

Update orbit to the latest version:

```bash
orbit update
```

Use `--check` to only verify the current version without installing:

```bash
orbit update --check
```

Use `--version v0.0.4` to install a specific version.

## Configuration

Edit `~/.config/.orbit/config.yaml` to define your projects:

```yaml
projects:
  my-webapp:
    ssh:
      host: example.com
      user: deployer
      auth:
        type: key
        key_path: ~/.ssh/id_rsa
    local:
      before:
        - npm install
        - npm run build
      after:
        - npm test
    upload:
      - source: ./dist
        destination: /var/www/app
    remote:
      commands:
        - cp -r /tmp/dist/* /var/www/app/
```

## Configuration Validation

Run `orbit validate` to check your configuration for common issues:

```bash
orbit validate
```

This will warn you about:
- Projects with `before` but no `after` section
- Empty `local` sections

These warnings don't block execution, but they may indicate misconfigurations.

## Commands

- `orbit init` — Create a new config file
- `orbit list` — Show all configured projects
- `orbit validate` — Validate config and report warnings
- `orbit update` — Update orbit to the latest version
- `orbit update --check` — Check for updates without installing
- `orbit run <name>` — Execute a project
- `orbit run <name> --dry-run` — Preview without executing
- `orbit version` — Show version

## SSH Aliases

You can use SSH aliases instead of specifying host/key manually. Just set up `~/.ssh/config`:

```
Host myserver
  HostName example.com
  User deployer
  IdentityFile ~/.ssh/id_ed25519
```

Then in your project:

```yaml
ssh:
  alias: myserver
  user: deployer
  auth:
    type: key
```

## Requirements

- Go 1.25+
- SSH access to your servers
- `~/.local/bin` in your PATH (or any directory you choose)

## Building

```bash
# Build for current platform
go build -o orbit .

# Cross-compile for multiple platforms
VERSION=0.0.1 ./scripts/build_all.sh
```

Binaries go in the `build/` directory.

## Security

- Don't commit config files with credentials to git
- Use SSH keys instead of passwords when possible
- Set restrictive permissions on your config: `chmod 600 ~/.config/orbit/config.yaml`
- SSH host key verification is disabled (`ssh.InsecureIgnoreHostKey`) — use within trusted networks

## License

MIT
