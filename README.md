🚀 Remote CLI Automation Tool

A powerful CLI application that simplifies remote server workflows by combining SSH access, file transfers, and local/remote command execution into configurable, reusable projects.

Built with Go (Golang) and powered by Cobra CLI for a fast, scalable, and maintainable command-line experience.

✨ Features
🔐 Connect to remote servers via SSH:
  - Username & password
  - SSH alias (from SSH config)
  - Private key with passphrase
⚡ Execute commands remotely
📦 Upload files to remote servers (optional)
🛠 Run local commands before and after remote execution
📂 Support for running local commands in custom directories
🧩 Create reusable projects
💾 Persistent configuration using YAML
🔄 Automate workflows like:
  Build locally → (optional upload) → deploy remotely

📌 Table of Contents
- Quick Start
- Requirements
- Project Structure
- Configuration
- SSH Alias Guide
- Examples
- Usage
- How It Works
- Authentication Methods
- Troubleshooting
- Roadmap
- Contributing
- License

🚀 Quick Start

# install

go install github.com/cristiangonsevi/orbit@latest

# ensure Go bin is in PATH
export PATH="$HOME/.local/bin:$PATH"

# initialize config (creates a YAML template)
orbit init

# run your first project
orbit run my-app

📍 Binary location:

~/.local/bin/orbit

📍 Config file location:

~/.config/ssh-deployer/config.yaml

📁 Requirements
- Go 1.20+ installed
- SSH access to remote hosts
- `~/.local/bin` added to your PATH
- Optional: SSH config with aliases defined in `~/.ssh/config`

📁 Project Structure

The project follows a clean and scalable structure using Cobra commands:

.
├── cmd/                # CLI commands (Cobra)
│   ├── root.go
│   ├── run.go
│   └── list.go
├── internal/           # Core application logic
│   ├── ssh/            # SSH connection & execution
│   ├── config/         # YAML parsing & persistence
│   ├── executor/       # Workflow orchestration
│   └── uploader/       # File transfer logic
├── pkg/                # Reusable public packages (optional)
├── configs/            # Example or default configs
├── main.go             # Entry point
└── README.md

📁 Configuration

Projects are stored in YAML format and allow you to define full workflows.

> **Note:** The `upload` section is optional. If omitted, the workflow will skip file transfer and only execute local and remote commands.

> **Note:** Remote commands are executed sequentially in the same shell session, so context is preserved. If you run `cd /folder` and then `ls`, the `ls` command runs inside `/folder`.

> **Note:** Local commands run in the current working directory by default. Use `working_dir` to change it.

### YAML structure

```yaml
projects:
  project-name:
    ssh:
      host: example.com        # required unless using ssh alias
      user: deployer
      auth:
        type: key              # key, password
        key_path: ~/.ssh/id_rsa
        passphrase: secret
        # password: secret      # only for password auth
    local:
      working_dir: ./app
      before:
        - npm install
        - npm run build
      after:
        - echo "Done"
    upload:
      - source: ./app/dist
        destination: /srv/www/app
    remote:
      commands:
        - cd /srv/www/app
        - npm ci --omit=dev
        - sudo systemctl restart app.service
```

### Fields
- `projects`: top-level map of project definitions
- `ssh.host`: remote host or omit when using `ssh.alias`
- `ssh.alias`: SSH config alias from `~/.ssh/config`
- `ssh.user`: remote SSH user
- `ssh.auth.type`: `key` or `password`
- `ssh.auth.key_path`: path to private key (only for key auth)
- `ssh.auth.passphrase`: passphrase for the private key
- `ssh.auth.password`: password-based SSH auth
- `local.working_dir`: local directory for `before`/`after` commands
- `local.before`: commands executed before upload/remote steps
- `local.after`: commands executed after remote steps
- `upload`: optional array of files/directories to send
- `remote.commands`: sequential commands executed on the remote host

### SSH Alias Guide

If you use an SSH alias, define it in `~/.ssh/config`:

```text
Host backend-prod-alias
  HostName example.com
  User produser
  IdentityFile ~/.ssh/id_ed25519
```

Then the project can omit `host` and `key_path`:

```yaml
ssh:
  alias: backend-prod-alias
  user: produser
  auth:
    type: key
    passphrase: supersecurepass
```

## Examples

### Minimal project (no upload)

```yaml
projects:
  simple-app:
    ssh:
      host: example.com
      user: deployer
      auth:
        type: key
        key_path: ~/.ssh/id_rsa
        passphrase: example-passphrase
    local:
      before:
        - echo "Starting deployment"
    remote:
      commands:
        - cd /srv/simple-app
        - ls
        - sudo systemctl restart simple-app.service
```

### Realistic Node.js deploy with upload

```yaml
projects:
  my-webapp:
    ssh:
      host: 192.168.1.100
      user: deployer
      auth:
        type: key
        key_path: ~/.ssh/id_rsa
        passphrase: my-secret-passphrase
    local:
      working_dir: ./webapp
      before:
        - echo "Building frontend..."
        - npm ci
        - npm run build
      after:
        - echo "Local build complete."
    upload:
      - source: ./webapp/dist
        destination: /srv/www/webapp
    remote:
      commands:
        - cd /srv/www/webapp
        - echo "Installing dependencies..."
        - npm ci --omit=dev
        - echo "Restarting service..."
        - sudo systemctl restart webapp.service
```

### Deploy using SSH alias and passphrase

```yaml
projects:
  prod-backend:
    ssh:
      alias: backend-prod-alias
      user: produser
      auth:
        type: key
        passphrase: supersecurepass
    local:
      working_dir: ./backend
      before:
        - echo "Running Go build..."
        - go build -o backend-app .
      after:
        - echo "Local build done."
    upload:
      - source: ./backend/backend-app
        destination: /opt/backend/backend-app
    remote:
      commands:
        - chmod +x /opt/backend/backend-app
        - sudo systemctl restart backend.service
```

### Deploy using username and password

```yaml
projects:
  staging-api:
    ssh:
      host: staging.example.com
      user: apiuser
      auth:
        type: password
        password: mypassword123
    local:
      working_dir: ./api
      before:
        - echo "Running tests..."
        - pytest
        - echo "Packaging app..."
        - tar czf api.tar.gz .
      after:
        - rm api.tar.gz
    upload:
      - source: ./api/api.tar.gz
        destination: /tmp/api.tar.gz
    remote:
      commands:
        - tar xzf /tmp/api.tar.gz -C /srv/api
        - sudo systemctl restart api.service
```

## Usage

```bash
orbit init
orbit list
orbit run <project-name>
orbit run <project-name> --dry-run
```

### Common flags
- `--dry-run`: validate the project without executing commands
- `--config <path>`: use a custom config file
- `--verbose`: enable detailed output

## How It Works
- Load project from YAML
- Execute local `before` commands
- Connect to the remote server via SSH
- Upload files if configured
- Execute remote commands sequentially
- Execute local `after` commands

## Authentication Methods
Supported SSH authentication methods:
- Password
- SSH key + passphrase
- SSH alias from `~/.ssh/config`

## Troubleshooting

### Common issues
- `Permission denied`: check SSH key permissions and user access
- `Host unreachable`: verify network connectivity and host address
- `Upload failed`: confirm remote destination permissions and path
- `Command failed`: remote commands run sequentially, so an earlier failure stops the workflow

### Tips
- Use `orbit run <project-name> --dry-run` to validate config before deployment
- Keep passwords out of YAML when possible; prefer SSH keys or environment-based secrets
- Ensure your SSH alias works with `ssh backend-prod-alias` before using it in the project

## Project-Based Design
Each project represents an application or environment:
- Independent configuration
- Reusable workflows
- Easy switching between servers/environments

## Roadmap
- Parallel execution across multiple servers
- Environment variables support
- Secrets management
- Plugin system

## Contributing
Contributions are welcome! Feel free to open issues or submit pull requests.

## License
MIT License
