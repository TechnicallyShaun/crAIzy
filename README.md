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
- An AI provider API key (OpenAI, Anthropic, etc.)

### From Source

```bash
git clone https://github.com/yourusername/crAIzy.git
cd crAIzy
go build -o crazy ./cmd/crazy
sudo mv crazy /usr/local/bin/
```

## Usage

### Quick Start

```bash
# Start the crAIzy dashboard
crazy

# Spawn a new AI agent
crazy new "implement user authentication"

# List all running agents
crazy ls

# Attach to an agent session
crazy attach 1

# Check agent status
crazy status
```

## Commands

| Command | Description |
|---------|-------------|
| `crazy` | Launch the interactive dashboard |
| `crazy new <task>` | Spawn a new AI agent with a task |
| `crazy ls` | List all active agent sessions |
| `crazy attach <id>` | Attach to a specific agent session |
| `crazy kill <id>` | Terminate an agent session |
| `crazy status` | Show status of all agents |
| `crazy pause <id>` | Pause an agent |
| `crazy resume <id>` | Resume a paused agent |
| `crazy logs <id>` | View agent logs |

## Configuration

Create `~/.config/crAIzy/config.yaml`:

```yaml
ai:
  provider: anthropic  # or openai, ollama
  model: claude-sonnet-4-20250514

tmux:
  prefix: crAIzy

dashboard:
  refresh_rate: 1s
  theme: dark
```

## Contributing

Contributions are welcome! Please read our contributing guidelines before submitting PRs.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

<p align="center">
  <sub>Built with 🧠 and ☕ by developers who talk to their terminal</sub>
</p>
