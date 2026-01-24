# TMUX Status Bar Customization

Notes on what's possible with the tmux status bar when porting into agent sessions.

## Current Implementation

Status bar is configured in `internal/infra/tmux_client.go` via `configureStatusBar()`.

Current layout:
```
 crAIzy | session-name |     window     | Detach: Ctrl+B, D | 21:26
```

## Available Variables

Useful tmux format variables that could be displayed:

| Variable | Description | Example |
|----------|-------------|---------|
| `#{session_name}` | Session identifier | `craizy-proj-claude` |
| `#{pane_current_path}` | Current working directory | `/home/user/project` |
| `#{pane_pid}` | Process ID of pane | `12345` |
| `#H` | Hostname | `dev-machine` |
| `#{client_width}x#{client_height}` | Terminal dimensions | `120x40` |
| `#{window_name}` | Current window name | `claude` |
| `#{pane_current_command}` | Running command | `claude` |
| `#(command)` | Output of shell command | Custom data |

## Ideas for Future Enhancement

### Agent-Specific Info
- Show agent type prominently (claude, aider, etc.)
- Display agent name separately from session ID
- Show project/repo name

### Status Indicators
- Git branch: `#(git -C #{pane_current_path} branch --show-current 2>/dev/null)`
- Git dirty status: custom script checking for uncommitted changes
- Agent activity indicator (active/idle)

### Resource Info
- Memory usage of agent process
- Token count if available from agent CLI
- Session duration/uptime

### Navigation Hints
- `Ctrl+B, D` - Detach (return to crAIzy)
- `Ctrl+B, [` - Scroll mode (navigate output)
- `Ctrl+B, c` - New window
- `Ctrl+B, n/p` - Next/prev window

### Theming
- Match status bar colors to agent type (purple for claude, green for aider, etc.)
- Dark/light mode support
- User-configurable colors

## Implementation Notes

All options are set per-session using:
```go
exec.Command("tmux", "set-option", "-t", sessionID, "option-name", "value").Run()
```

Shell commands in status bar via `#(command)` are evaluated periodically by tmux (default every 15 seconds, configurable with `status-interval`).
