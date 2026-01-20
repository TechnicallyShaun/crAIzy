<p align="center">
  <img src="https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go"/>
  <img src="https://img.shields.io/badge/tmux-1BB91F?style=for-the-badge&logo=tmux&logoColor=white" alt="tmux"/>
  <img src="https://img.shields.io/badge/AI-FF6F61?style=for-the-badge&logo=openai&logoColor=white" alt="AI"/>
</p>

<h1 align="center">
  🤖 crAIzy
</h1>

<p align="center">
  <strong>tmux management tool for orchestrating AI agents</strong>
</p>

<p align="center">
  <a href="#features">Features</a> •
  <a href="#installation">Installation</a> •
  <a href="#usage">Usage</a> •
  <a href="#commands">Commands</a> •
  <a href="PLAN.md">Development Plan</a> •
  <a href="VISION.md">Vision</a>
</p>

---

## What is crAIzy?

**crAIzy** is a tmux management tool that allows the orchestration of AI agents. It provides an easy-to-use TUI (Terminal User Interface) for orchestrating bash/CLI tools, primarily designed for AI agent management.

Think of it as a window manager for AI—spin up agents in parallel using git worktrees, switch between them, manage code changes, and let them collaborate on different parts of your project simultaneously. crAIzy handles the entire workflow from agent spawning to PR creation.

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
- **📊 Status Dashboard** — Real-time overview of all running agents with TUI interface
- **🌲 Git Worktree Support** — Agents work in isolated git worktrees for parallel development
- **🔍 Change Detection** — Automatically detect uncommitted changes in worktrees
- **🚀 Automated Workflow** — Push changes and open PRs directly from the interface
- **🔄 Auto-Sync** — Periodically fetch and pull to keep worktrees up-to-date

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

### Complete User Journey

crAIzy provides a streamlined workflow for orchestrating AI agents to work on your code:

1. **Install crAIzy** (see [Installation](#installation) above)

2. **Initialize a project:**
   ```bash
   craizy init myproject
   ```
   This creates a project directory with `.craizy/` configuration.

3. **Navigate to your project:**
   ```bash
   cd myproject
   ```

4. **Launch the dashboard:**
   ```bash
   craizy start
   ```
   This starts the interactive TUI dashboard powered by tmux.

5. **Spawn a new AI agent:**
   - Press `n` in the dashboard
   - A modal appears asking which AI to launch
   - Select your desired AI from the configured options

6. **Interact with the AI:**
   - Press `Enter` to attach to the AI session
   - The AI agent will work in an isolated git worktree
   - Ask the AI to make changes to your source code
   - The AI can clone repos, make edits, and work in parallel with other agents

7. **Monitor git changes:**
   - crAIzy automatically detects uncommitted changes in each worktree
   - Visual indicators show which agents have pending changes
   - Each agent works in its own branch via git worktree

8. **Push changes:**
   - Press `p` to push changes from the worktree
   - Changes are pushed to the origin repository on the agent's branch

9. **Create a Pull Request:**
   - Press `r` to open a PR
   - The PR is created from the agent's branch into the main branch
   - crAIzy handles the entire git workflow

10. **Stay synchronized:**
    - crAIzy periodically fetches and pulls new changes from origin
    - New worktrees are always created from an up-to-date state
    - This ensures all agents work with the latest code

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

This launches an interactive tmux-based TUI dashboard where you can:
- Press **N** to spawn a new AI instance (opens modal to select AI)
- Press **Enter** to attach to and interact with an AI session
- Use number keys to select specific AI sessions
- View AI output and git status in real-time
- Monitor uncommitted changes across all worktrees
- Push changes and create PRs with keyboard shortcuts
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
