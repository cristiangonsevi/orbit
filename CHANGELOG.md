# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.0.1] - 2025-05-01

### Added
- `orbit init` - Initialize configuration file
- `orbit list` - List available projects
- `orbit run <project>` - Execute project workflow
- `orbit run --dry-run` - Validate without executing
- `orbit version` - Show version info
- `orbit completion` - Shell autocompletion (bash, zsh, fish, powershell)
- SSH authentication via password or key-based auth
- SSH alias support (reads ~/.ssh/config)
- Passphrase support for encrypted keys
- SSH agent integration
- SCP file/directory upload
- Local before/after commands execution
- Custom working directory support
- Colored CLI output with spinners
- `--quiet` flag for CI/CD environments
- `--verbose` flag for debugging
- Cross-platform build scripts (linux/darwin, amd64/arm64)
- Install script for easy distribution
- Comprehensive README documentation
- Unit tests for config package

### Fixed
- Error handling improvements
- Usage silence for commands
- Loading spinner formatting
