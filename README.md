<p align="center">
  <img src="https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go"/>
  <img src="https://img.shields.io/badge/tmux-1BB91F?style=for-the-badge&logo=tmux&logoColor=white" alt="tmux"/>
  <img src="https://img.shields.io/badge/AI-FF6F61?style=for-the-badge&logo=openai&logoColor=white" alt="AI"/>
</p>

<h1 align="center">
  🤖 crAIzy
</h1>

<p align="center">
  <strong>AI-powered terminal multiplexer for orchestrating intelligent agents</strong>
</p>

<p align="center">
  <a href="#features">Features</a> •
  <a href="#installation">Installation</a> •
  <a href="#usage">Usage</a> •
  <a href="#commands">Commands</a> •
  <a href="VISION.md">Vision</a>
</p>

---

## What is crAIzy?

**crAIzy** is a CLI tool that combines the power of tmux with AI agents, giving you a unified interface to spawn, manage, and orchestrate multiple AI sessions from your terminal.

Think of it as a window manager for AI—spin up agents, switch between them, and let them work on different parts of your project simultaneously.

```
┌──────────────────────────────────────────────────────────────┐
│  crAIzy                                              v0.1.0  │
├──────────────────────────────────────────────────────────────┤
│  ┌─────────────────────┐  ┌─────────────────────┐           │
│  │ [1] feature/auth    │  │ [2] bugfix/api      │           │
│  │ ████████░░ 80%      │  │ ██████████ Done     │           │
│  │ Implementing JWT... │  │ Ready for review    │           │
│  └─────────────────────┘  └─────────────────────┘           │
│  ┌─────────────────────┐  ┌─────────────────────┐           │
│  │ [3] refactor/db     │  │ [4] docs/readme     │           │
│  │ ████░░░░░░ 40%      │  │ ░░░░░░░░░░ Queued   │           │
│  │ Analyzing schema... │  │ Waiting...          │           │
│  └─────────────────────┘  └─────────────────────┘           │
└──────────────────────────────────────────────────────────────┘
```

## Features

- **🪟 Multi-Agent Management** — Run multiple AI agents in parallel tmux sessions
- **⚡ Quick Spawn** — Fire up AI agents with a single command
- **🔄 Session Switching** — Seamlessly jump between active AI sessions
- **📊 Status Dashboard** — Real-time overview of all running agents
- **🎯 Task Assignment** — Direct agents to specific tasks or files
- **🔗 Agent Coordination** — Let agents collaborate on complex tasks

## Installation

### Prerequisites

- Go 1.21+
- tmux 3.0+
- Linux/Ubuntu (or macOS with tmux)

### Using go install

```bash
go install github.com/TechnicallyShaun/crAIzy/cmd/craizy@latest
```

### From Source

```bash
git clone https://github.com/TechnicallyShaun/crAIzy.git
cd crAIzy
make build
sudo cp bin/craizy /usr/local/bin/
```

Or use `make install` to build and install in one step:

```bash
make install
```

### From Release

Download the latest binary for your platform from the [releases page](https://github.com/TechnicallyShaun/crAIzy/releases):

```bash
# Linux AMD64
wget https://github.com/TechnicallyShaun/crAIzy/releases/latest/download/craizy-linux-amd64
chmod +x craizy-linux-amd64
sudo mv craizy-linux-amd64 /usr/local/bin/craizy

# macOS ARM64 (Apple Silicon)
wget https://github.com/TechnicallyShaun/crAIzy/releases/latest/download/craizy-darwin-arm64
chmod +x craizy-darwin-arm64
sudo mv craizy-darwin-arm64 /usr/local/bin/craizy
```

## Usage

### Quick Start

```bash
# Initialize a new crAIzy project
craizy init my-project
cd my-project

# Start the dashboard
craizy start
```

### Initialize a Project

Create a new crAIzy project in a directory:

```bash
craizy init <project-name>
```

This creates:
- A directory named `<project-name>`
- A `.craizy/` subdirectory with configuration files
- `config.yaml` - Project configuration
- `ais.yaml` - AI definitions and commands

### Start the Dashboard

From within a crAIzy project directory:

```bash
craizy start
```

This launches an interactive tmux-based dashboard where you can:
- Press **N** to spawn a new AI instance
- Use number keys to select and attach to AI sessions
- View AI output in the preview window
- Manage multiple AI sessions simultaneously

### Configuration

Edit `.craizy/ais.yaml` to customize available AI options:

```yaml
ais:
  - name: GPT-4
    command: openai-cli chat --model gpt-4
    options:
      api_key: $OPENAI_API_KEY
  
  - name: Claude
    command: anthropic-cli chat --model claude-3-opus
    options:
      api_key: $ANTHROPIC_API_KEY
  
  - name: Local LLaMA
    command: ollama run llama2
```

Each AI entry specifies:
- **name**: Display name for the AI
- **command**: Shell command to run the AI
- **options**: Environment variables or configuration (optional)

## Commands

| Command | Description |
|---------|-------------|
| `craizy init <name>` | Initialize a new crAIzy project |
| `craizy start` | Start the interactive dashboard |
| `craizy version` | Show version information |
| `craizy help` | Display help message |

## Building and Testing

### Build

```bash
make build
```

The binary will be created at `bin/craizy`.

### Run Tests

```bash
make test
```

### Run Linter

```bash
make lint
```

### Generate Coverage Report

```bash
make coverage
```

## Configuration

Configuration is stored in `.craizy/config.yaml`:

```yaml
project_name: my-project
```

AI definitions are in `.craizy/ais.yaml`:

```yaml
ais:
  - name: GPT-4
    command: openai-cli chat --model gpt-4
    options:
      api_key: $OPENAI_API_KEY
```

## CI/CD

This project includes GitHub Actions workflows for:

- **Build & Test**: Runs on every push and PR
- **Linting**: Code quality checks with golangci-lint
- **Release**: Automated releases with multi-platform binaries

To create a release, push a tag:

```bash
git tag v0.1.0
git push origin v0.1.0
```

## Development

### Project Structure

```
crAIzy/
├── cmd/
│   └── craizy/          # Main entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── tmux/            # Tmux session management
│   └── ui/              # Dashboard UI
├── .github/
│   └── workflows/       # CI/CD pipelines
├── Makefile             # Build automation
└── .golangci.yml        # Linter configuration
```

### Adding Tests

Tests should be placed alongside the code they test with `_test.go` suffix:

```bash
# Run all tests
make test

# Run specific package tests
go test -v ./internal/config
```

## Contributing

Contributions are welcome! Please read our contributing guidelines before submitting PRs.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

<p align="center">
  <sub>Built with 🧠 and ☕ by developers who talk to their terminal</sub>
</p>
