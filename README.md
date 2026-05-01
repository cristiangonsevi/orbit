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

## Installation

Download a binary for your platform from the releases page, or use the install script:

```bash
curl -fsSL https://raw.githubusercontent.com/cristiangonsevi/orbit/refs/heads/master/scripts/install.sh | bash
```

The binary installs to `~/.local/bin/orbit` by default.

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
    upload:
      - source: ./dist
        destination: /var/www/app
    remote:
      commands:
        - cp -r /tmp/dist/* /var/www/app/
```

## Commands

- `orbit init` — Create a new config file
- `orbit list` — Show all configured projects
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

## License

MIT
