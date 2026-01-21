<p align="center">
  <img src="https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go"/>
  <img src="https://img.shields.io/badge/tmux-1BB91F?style=for-the-badge&logo=tmux&logoColor=white" alt="tmux"/>
  <img src="https://img.shields.io/badge/Bubble%20Tea-F05032?style=for-the-badge&logo=tea&logoColor=white" alt="Bubble Tea"/>
  <img src="https://img.shields.io/badge/AI-FF6F61?style=for-the-badge&logo=openai&logoColor=white" alt="AI"/>
</p>

<h1 align="center">
  🤖 crAIzy
</h1>

<p align="center">
  <strong>tmux session manager for orchestrating AI agents</strong>
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

**crAIzy** is a TUI (Terminal User Interface) that acts as a mission control for AI agents. It orchestrates parallel development by managing distinct **tmux sessions** for each agent, giving them isolated git worktrees to code in.

It uses the **Bubble Tea** framework to provide a rich, interactive dashboard.

## The Experience

### 1. The Dashboard (Mission Control)
The main entry point is a split-screen TUI that allows you to monitor all active agents at a glance.

```
┌ crAIzy ──────────────────────────────────────────────────────────────────┐
│  ACTIVE AGENTS (3)    │  PREVIEW: Feature/Auth                           │
│ ──────────────────────│ ──────────────────────────────────────────────── │
│ > [1] Feature/Auth    │  > Claude: I have updated auth.go                │
│   ● Active (Claude)   │  > User: Run the tests please.                   │
│   🌿 feat/login       │  > Claude: Running go test ./...                 │
│                       │  PASS: TestLogin (0.02s)                         │
│   [2] Bugfix/API      │  PASS: TestLogout (0.01s)                        │
│   ○ Idle (GPT-4)      │  ok      github.com/app/auth     0.435s          │
│   🌿 fix/api-timeout  │                                                  │
│                       │  > Claude: Tests passed. Ready to push?          │
│   [3] Docs/Readme     │  _                                               │
│   ○ Idle (Aider)      │                                                  │
│   🌿 docs/update      │                                                  │
│                       │                                                  │
│                       │                                                  │
└───────────────────────┴──────────────────────────────────────────────────┘
  [n] New Agent   [↑/↓] Navigate   [Enter] Attach   [q] Quit Dashboard
```

### 2. The HUD (In-Session)
When you attach to an agent (Press `Enter`), you are switched to that agent's dedicated tmux session. crAIzy automatically splits the window to keep a persistent **HUD** at the top, ensuring you always have context and know the controls.

```
┌──────────────────────────────────────────────────────────────────────────┐
│ 🤖 AGENT: Feature/Auth  |  🌿 BRANCH: feat/login                         │
│ 🎮 CONTROLS: [Ctrl+b d] Detach to Dashboard  |  [Ctrl+b s] Session List  │
├──────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  > Claude: I'm ready to help. What's the task?                           │
│                                                                          │
│  > User: Refactor the login handler.                                     │
│                                                                          │
│  > Claude: On it. Checking files...                                      │
│  _                                                                       │
│                                                                          │
└──────────────────────────────────────────────────────────────────────────┘
```

## Features

- **🖥️ Rich TUI Dashboard** — Built with Bubble Tea. Navigate with arrow keys, no complex CLI commands.
- **⚡ Hotkey-Driven** — Press `n` for a modal to spawn agents. No `Enter` required.
- **🪟 Session Isolation** — Each agent runs in its own full **tmux session**, not just a window.
- **👀 Live Previews** — The dashboard polls and displays the live output of any selected agent.
- **🧭 Persistent HUD** — Never get lost in a terminal again. Every agent session includes a read-only top bar with navigation help.
- **🌲 Git Worktree Support** — Agents work in isolated git worktrees for parallel development.
- **🔄 Native Tmux Navigation** — Compatible with standard tmux controls (`Ctrl+b s`, `Ctrl+b d`).

## Installation

### Prerequisites

- Go 1.21+
- tmux 3.0+
- Linux/Ubuntu (or macOS with tmux)

### Using go install

```bash
go install github.com/TechnicallyShaun/crAIzy/cmd/craizy@latest
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

### Workflow

1.  **Launch Dashboard:** Run `craizy start`. You see the agent list (empty initially).
2.  **Spawn Agent:** Press `n`. A modal pops up. Select "Claude" (or your configured agent) using arrow keys and press `Enter`.
3.  **Attach:** The dashboard creates a new tmux session and worktree. Highlight the new agent and press `Enter`.
4.  **Interact:** You are now in the agent's session. The top HUD shows you how to leave (`Ctrl+b d`).
5.  **Detach:** Press `Ctrl+b d`. You are instantly back at the Dashboard.

### Configuration

Configuration is stored in `.craizy/config.yaml`:

```yaml
project_name: my-project
```

AI agent definitions are in `.craizy/agents.yaml`:

```yaml
agents:
  - name: Claude
    command: claude --dangerously-skip-permissions
  - name: Aider
    command: aider
```

## Commands

| Command | Description |
|---------|-------------|
| `craizy init <name>` | Initialize a new crAIzy project |
| `craizy start` | Start the interactive TUI dashboard |
| `craizy agent add` | Add a new AI agent configuration |
| `craizy agent list` | List configured agents |

## Development

### Tech Stack
*   **Language:** Go
*   **TUI Framework:** [Bubble Tea](https://github.com/charmbracelet/bubbletea) (The Elm Architecture)
*   **Styling:** [Lip Gloss](https://github.com/charmbracelet/lipgloss)
*   **Multiplexer:** tmux (via CLI wrapper)

### Project Structure

```
crAIzy/
├── cmd/
│   ├── craizy/          # Main entry point
│   └── hud/             # The lightweight binary for the session top-bar
├── internal/
│   ├── config/          # Configuration management
│   ├── tmux/            # Tmux session/window orchestration
│   └── tui/             # Bubble Tea models and views
│       ├── dashboard/   # Main dashboard logic
│       └── hud/         # HUD display logic
└── .github/
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.