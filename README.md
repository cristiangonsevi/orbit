🚀 Remote CLI Automation Tool

A powerful CLI application that simplifies remote server workflows by combining SSH access, file transfers, and local/remote command execution into configurable, reusable projects.

Built with Go (Golang) and powered by Cobra CLI for a fast, scalable, and maintainable command-line experience.

✨ Features
🔐 Connect to remote servers via SSH:
Username & password
SSH alias (from SSH config)
Private key with passphrase
⚡ Execute commands remotely
📦 Upload files to remote servers (optional)
🛠 Run local commands before and after remote execution
📂 Support for running local commands in custom directories
🧩 Create reusable projects
💾 Persistent configuration using YAML
🔄 Automate workflows like:
Build locally → (optional upload) → deploy remotely
🚀 Quick Start
# install

go install github.com/cristiangonsevi/orbit@latest

# ensure Go bin is in PATH (common setup)
export PATH="$HOME/.local/bin:$PATH"


# initialize config (creates a YAML template)
orbit init


# run your first profile
orbit run my-app


📍 Binary location:


~/.local/bin/orbit


📍 Config file location:

~/.config/ssh-deployer/config.yaml

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

💡 Remote commands are executed sequentially in the same shell session, so context is preserved. For example, if you run cd folder and then ls, the ls will be executed inside that folder.

💡 Local commands run in the current working directory by default, but you can override this using working_dir.



🧪 Example: Realistic Deploy of a Node.js App (SSH Key)
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
        - echo "Switching to app directory..."
        - cd /srv/www/webapp
        - echo "Installing dependencies..."
        - npm ci --omit=dev
        - echo "Restarting service..."
        - sudo systemctl restart webapp.service


🧪 Example: Deploy using SSH Alias & Passphrase

projects:
  prod-backend:
    ssh:
      alias: backend-prod-alias   # defined in ~/.ssh/config
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
        - echo "Setting permissions..."
        - chmod +x /opt/backend/backend-app
        - echo "Restarting backend service..."
        - sudo systemctl restart backend.service


🧪 Example: Deploy using Username & Password
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
        - echo "Unpacking and deploying API..."
        - tar xzf /tmp/api.tar.gz -C /srv/api
        - echo "Restarting API service..."
        - sudo systemctl restart api.service

🚀 Usage



Run a project:

orbit run node-app




Other examples:

orbit list
orbit run node-app --dry-run

⚙️ How It Works
Load project from YAML
Execute local "before" commands (in configured working directory)
Connect to the remote server via SSH
Upload files (if configured)
Run remote commands
Execute local "after" commands
🔒 Authentication Methods

Supported SSH authentication methods:

Password
SSH key + passphrase
SSH config alias (~/.ssh/config)
🧩 Project-Based Design

Each project represents an application or environment:

Independent configuration
Reusable workflows
Easy switching between servers/environments
📌 Roadmap
 Parallel execution across multiple servers
 Environment variables support
 Secrets management
 Plugin system
🤝 Contributing

Contributions are welcome! Feel free to open issues or submit pull requests.

📄 License

MIT License
