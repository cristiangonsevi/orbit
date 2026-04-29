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

💡 The upload section is optional — if omitted, the workflow will skip file transfer.

💡 Local commands run in the current working directory by default, but you can override this using working_dir.


🧪 Example: Deploy a Node.js App (SSH Key)
projects:
  node-app:
    ssh:
      host: example.com
      user: root
      auth:
        type: key
        key_path: ~/.ssh/id_rsa
        passphrase: your-passphrase

    local:
      working_dir: ./frontend   # optional, defaults to current directory
      before:
        - npm install
        - npm run build
      after:
        - echo "Deployment finished"

    upload: # optional
      - source: ./frontend/dist
        destination: /var/www/app

    remote:
      commands:
        - cd /var/www/app
        - npm install --production
        - pm2 restart app

🧪 Example: Deploy using Username & Password
projects:
  password-app:
    ssh:
      host: example.com
      user: myuser
      auth:
        type: password
        password: mypassword

    local:
      before:
        - echo "Preparando despliegue"

    remote:
      commands:
        - echo "¡Despliegue exitoso!"

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
