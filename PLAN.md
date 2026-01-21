# crAIzy Development Plan

## Overview

crAIzy is a **tmux session management tool** built with **Bubble Tea**. It acts as a GUI-like layer over terminal processes, specifically designed for orchestrating multiple AI agents in parallel.

## Core Architecture Shifts

1.  **Framework:** Move from standard CLI/`fmt` to **Bubble Tea** (Charm) for the Dashboard.
2.  **Session Model:** Move from "Windows in one session" to **"One Session Per Agent"**.
3.  **Visuals:**
    *   **Dashboard:** Interactive List (Left) + Live Preview (Right).
    *   **Agent Session:** Split Pane. Top 10% = Passive HUD. Bottom 90% = Active Agent.

## Implementation Roadmap

### Phase 1: The Bubble Tea Dashboard 🎨
**Goal:** Replace the static output with an interactive TUI.

- [ ] **Scaffold TUI:** Set up `tea.Model`, `tea.Cmd`, and `tea.Update` loop.
- [ ] **Layout:** Use `lipgloss` to create the Master-Detail (List vs Preview) split.
- [ ] **Agent List:** Implement a scrollable list of active sessions using `bubbles/list`.
- [ ] **Preview Pane:** Implement a `tea.Cmd` that runs `tmux capture-pane` every 2 seconds to update the preview view.
- [ ] **Help Bar:** Permanent footer using `bubbles/help` or static lipgloss style.

### Phase 2: Session Orchestration 🪟
**Goal:** Manage discrete tmux sessions for isolation.

- [ ] **Session Creator:** Update `tmux` package to spawn *new sessions* (`tmux new-session`) instead of windows.
- [ ] **Session Naming:** Standardize naming convention (e.g., `craizy-<project>-<agent-id>`).
- [ ] **Worktree Binding:** Ensure each session starts in its specific git worktree directory.
- [ ] **Switching:** Implement the `Enter` key action -> `tmux switch-client -t <target-session>`.

### Phase 3: The HUD (Heads Up Display) 🧭
**Goal:** Ensure the user always has context inside an agent session.

- [ ] **HUD Binary:** Create a small sub-command `craizy hud` that renders the passive info bar.
- [ ] **Split Logic:** When spawning an agent:
    1. Create Session.
    2. Split Window vertically (Top 10% / Bottom 90%).
    3. Top Pane: Run `craizy hud --agent="Claude" --branch="feature/x"`.
    4. Bottom Pane: Run the actual Agent command.
- [ ] **Focus:** Ensure cursor/focus defaults to the Bottom Pane.

### Phase 4: Modals & Interaction ⌨️
**Goal:** "No-Enter" hotkey workflows.

- [ ] **Agent Spawner Modal:**
    *   Press `n` on Dashboard.
    *   Render a centered box (z-index overlay).
    *   List configured AIs from `agents.yaml`.
    *   Arrow keys to select, `Enter` to confirm.
- [ ] **Hotkeys:**
    *   `q`: Quit dashboard.
    *   `↑/↓`: Navigate list.
    *   `Enter`: Attach to session.

### Phase 5: Git Integration (Existing Goal)
- [ ] **Worktree Management:** (Remains same as original plan, just integrated into the new session starter).
- [ ] **Push/PR Actions:** Add hotkeys to the Dashboard (`p`, `r`) to trigger these actions on the *selected* agent.

## Technical Details

### The "Preview" Loop
To show what's happening in an agent's session without switching to it:
1.  Dashboard `Update()` triggers a `Tick` every 2 seconds.
2.  `Tick` returns a `Cmd` that executes `tmux capture-pane -p -t <session_name> | tail -n 20`.
3.  The result updates the `content` string of the Preview model.

### The "HUD" implementation
The HUD is a dumb terminal program. It does not accept input.
Command: `craizy internal-hud --name="Claude" --branch="feat/login"`
Output:
```text
 [ AGENT: Claude ] ──────────────────────────────── [ BRANCH: feat/login ]
 Controls: [Ctrl+b d] Dashboard  [Ctrl+b s] Sessions  [Ctrl+b c] New Tab
```
It waits for signals (like window resize) to redraw, but otherwise just sleeps.

## Dependencies to Add
- `github.com/charmbracelet/bubbletea`
- `github.com/charmbracelet/bubbles`
- `github.com/charmbracelet/lipgloss`

## Success Metrics
- [ ] Dashboard feels like a native GUI app (instant response, no scrolling logs).
- [ ] User can switch contexts using standard `tmux` keys.
- [ ] User never feels "trapped" in a session.
- [ ] Preview window gives enough context to decide whether to switch.