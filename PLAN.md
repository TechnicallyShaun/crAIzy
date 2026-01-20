# crAIzy Development Plan

## Overview

crAIzy is a **tmux management tool for AI orchestration** that provides an easy-to-use TUI (Terminal User Interface) for managing AI agents working on code. The tool orchestrates bash/CLI tools (primarily AI agents) to enable parallel software development with automatic git worktree management, change detection, and PR workflow automation.

## Core Capabilities

### 1. AI Agent Orchestration
- Spawn multiple AI agents in parallel tmux sessions
- Each agent works in an isolated environment
- Seamless switching between agents via TUI
- Real-time monitoring of agent activity

### 2. Git Worktree Management
- **Automatic Worktree Creation**: Each AI agent gets its own git worktree
- **Isolated Branches**: Agents work on separate branches without conflicts
- **Parallel Development**: Multiple agents can modify code simultaneously
- **Repository Operations**: Agents can clone repos and make changes independently

### 3. Change Detection & Visualization
- **Real-time Monitoring**: Track uncommitted changes in each worktree
- **Visual Indicators**: Dashboard shows which agents have pending changes
- **Status Updates**: Git status displayed for each active worktree
- **Branch Tracking**: Monitor which branch each agent is working on

### 4. Automated Git Workflow
- **Push to Origin**: Single key press to push worktree changes
- **PR Creation**: Automated pull request opening from agent branches
- **Branch Management**: Automatic branch creation and cleanup
- **Merge Workflow**: Facilitate merging agent changes to main branch

### 5. Synchronization
- **Periodic Fetch**: Automatically fetch latest changes from origin
- **Auto-Pull**: Keep main branch up-to-date
- **Fresh Worktrees**: New worktrees always start from current main
- **Conflict Prevention**: Reduce merge conflicts through regular syncing

## User Journey

### Standard Workflow

```
┌─────────────────────────────────────────────────────────────┐
│                  crAIzy User Journey                        │
└─────────────────────────────────────────────────────────────┘

1. Install crAIzy
   └─> One-time setup

2. craizy init myproject
   └─> Creates project folder with .craizy/ configuration

3. cd myproject
   └─> Navigate into project

4. craizy start
   └─> Launch TUI dashboard

5. Press 'n' in dashboard
   └─> Modal appears: "Which AI would you like to launch?"
   └─> Select from configured AIs (GPT-4, Claude, Local LLaMA, etc.)

6. Press Enter to attach
   └─> Enter the AI's tmux session
   └─> Agent is now active in its own git worktree

7. Interact with AI
   └─> Ask AI to clone repos, make changes, write code
   └─> AI works in isolated worktree on dedicated branch
   └─> Multiple agents can work in parallel

8. Visual feedback
   └─> Dashboard shows uncommitted changes indicator
   └─> Git status displayed for each worktree
   └─> Real-time updates as agents modify files

9. Push changes
   └─> Press 'p' to push to origin
   └─> Changes from worktree → remote branch

10. Open Pull Request
    └─> Press 'r' to create PR
    └─> PR: agent-branch → main
    └─> Automated description generation

11. Continuous sync
    └─> crAIzy periodically fetches origin
    └─> Main branch stays current
    └─> New worktrees use latest code
```

## Architecture

### Components

```
┌──────────────────────────────────────────────────────────┐
│                      crAIzy Stack                        │
├──────────────────────────────────────────────────────────┤
│                                                          │
│  ┌─────────────────────────────────────────────────┐   │
│  │              TUI Dashboard (Go)                  │   │
│  │  - Keyboard navigation                           │   │
│  │  - Real-time updates                            │   │
│  │  - Git status display                           │   │
│  └─────────────────────────────────────────────────┘   │
│                          │                              │
│                          ▼                              │
│  ┌─────────────────────────────────────────────────┐   │
│  │            tmux Session Manager                  │   │
│  │  - Create/destroy sessions                       │   │
│  │  - Attach/detach handling                       │   │
│  │  - Session state tracking                       │   │
│  └─────────────────────────────────────────────────┘   │
│                          │                              │
│                          ▼                              │
│  ┌─────────────────────────────────────────────────┐   │
│  │          Git Worktree Manager                    │   │
│  │  - Create worktrees per agent                   │   │
│  │  - Monitor git status                           │   │
│  │  - Handle push/PR operations                    │   │
│  │  - Periodic sync from origin                    │   │
│  └─────────────────────────────────────────────────┘   │
│                          │                              │
│                          ▼                              │
│  ┌─────────────────────────────────────────────────┐   │
│  │          AI Agent Processes                      │   │
│  │  - Each in own tmux session                     │   │
│  │  - Each in own git worktree                     │   │
│  │  - Bash/CLI tool execution                      │   │
│  └─────────────────────────────────────────────────┘   │
│                                                          │
└──────────────────────────────────────────────────────────┘
```

## Implementation Roadmap

### Phase 1: Foundation ✅ (Current)
- [x] Project initialization (`craizy init`)
- [x] Basic tmux session management
- [x] Configuration system (config.yaml, ais.yaml)
- [x] CLI interface
- [x] Basic dashboard display

### Phase 2: Git Worktree Integration (Next)
- [ ] Git worktree creation per agent
- [ ] Branch management for each worktree
- [ ] Git status monitoring
- [ ] Change detection in worktrees
- [ ] Visual indicators for uncommitted changes

### Phase 3: TUI Enhancement
- [ ] Interactive modal for AI selection
- [ ] Enhanced keyboard navigation
- [ ] Real-time git status display
- [ ] Split-pane layout for multi-agent view
- [ ] Color-coded status indicators
- [ ] Scrollable output windows

### Phase 4: Automated Workflows
- [ ] Push to origin functionality
- [ ] Automated PR creation
- [ ] PR description generation
- [ ] Keyboard shortcuts for git operations
- [ ] Branch cleanup after merge

### Phase 5: Synchronization
- [ ] Periodic git fetch from origin
- [ ] Auto-pull main branch updates
- [ ] Conflict detection and warnings
- [ ] Stale worktree notifications
- [ ] Automatic garbage collection

### Phase 6: Advanced Features
- [ ] Multi-repository support
- [ ] Agent collaboration (shared context)
- [ ] Custom workflow definitions
- [ ] Webhook integration for CI/CD
- [ ] Agent performance metrics
- [ ] Session replay/history

## Technical Details

### Git Worktree Strategy

Each AI agent operates in its own git worktree:

```bash
myproject/
├── .craizy/              # Configuration
├── .git/                 # Main git repository
├── main/                 # Main worktree (default)
└── worktrees/
    ├── agent-1-gpt4/    # Worktree for agent 1
    │   └── branch: feature/agent-1-task
    ├── agent-2-claude/  # Worktree for agent 2
    │   └── branch: feature/agent-2-task
    └── agent-3-llama/   # Worktree for agent 3
        └── branch: bugfix/agent-3-fix
```

Benefits:
- **Isolation**: Agents can't conflict with each other
- **Parallel Work**: Multiple changes happening simultaneously
- **Safety**: Main branch stays clean
- **Easy Cleanup**: Remove worktree = remove branch

### Change Detection

Monitor each worktree for:
- Modified files (`git status --porcelain`)
- Staged changes
- Untracked files
- Commits ahead of remote

Display in dashboard:
- 🔴 Red: Uncommitted changes
- 🟡 Yellow: Committed, not pushed
- 🟢 Green: Pushed, PR open
- ⚪ White: Clean worktree

### Keyboard Shortcuts (Planned)

| Key | Action |
|-----|--------|
| `n` | New AI agent |
| `Enter` | Attach to selected agent |
| `q` | Quit/detach |
| `p` | Push changes from current worktree |
| `r` | Create pull request |
| `s` | Show git status |
| `f` | Force fetch/sync |
| `k` | Kill selected agent |
| `↑/↓` | Navigate agent list |
| `Tab` | Switch focus (agents/preview) |

## Configuration

### .craizy/config.yaml
```yaml
project_name: myproject
git:
  remote: origin
  main_branch: main
  worktree_dir: ./worktrees
  auto_sync: true
  sync_interval: 300  # seconds
```

### .craizy/ais.yaml
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

## Success Metrics

- ✅ Multiple agents working in parallel without conflicts
- ✅ Seamless git worktree management
- ✅ Automatic change detection and visualization
- ✅ One-key push and PR creation
- ✅ Zero manual git command execution needed
- ✅ Agents stay synchronized with latest code

## Future Vision

Eventually, crAIzy will:
- Support team collaboration with shared agent pools
- Integrate with CI/CD for automated testing
- Provide analytics on agent productivity
- Enable custom agent behaviors and workflows
- Support advanced git operations (rebase, cherry-pick)
- Offer cloud-based agent orchestration

---

**crAIzy**: Orchestrating AI agents, one worktree at a time. 🤖🌲
