Epic: MVP

# Preview Pane

Dependencies: [spawn-agent-session.md](./spawn-agent-session.md)

## Description

The content area displays a live preview of the currently selected agent's tmux output. This allows users to monitor agent activity without porting in. The preview refreshes on a polling interval and shows the most recent output that fits within the pane.

When no agents exist, a branded welcome message is displayed.

## Stories

### Live Preview

As a user, I can see the output of my currently selected agent in the content area, refreshing automatically every 2 seconds.

#### Technical / Architecture

- Polling mechanism using Bubble Tea's `tea.Tick`:
  ```go
  const PreviewPollInterval = 2 * time.Second

  type PreviewTickMsg time.Time

  func (m *Model) pollPreview() tea.Cmd {
      return tea.Tick(PreviewPollInterval, func(t time.Time) tea.Msg {
          return PreviewTickMsg(t)
      })
  }
  ```

- On each tick, capture selected agent's pane:
  ```go
  func (s *AgentService) CaptureOutput(sessionID string, lines int) (string, error) {
      // Capture last N lines from tmux pane
      cmd := exec.Command("tmux", "capture-pane", "-t", sessionID, "-p", "-l", strconv.Itoa(lines))
      output, err := cmd.Output()
      return string(output), err
  }
  ```

- Add to `ITmuxClient` interface:
  ```go
  type ITmuxClient interface {
      // ... existing methods
      CapturePaneOutput(sessionID string, lines int) (string, error)
  }
  ```

- Polling lifecycle:
  1. Start polling when first agent is created
  2. On tick: capture output for selected agent only
  3. Stop polling when porting into agent
  4. Resume polling when detaching back to dashboard
  5. Stop polling when last agent is killed

- Message flow:
  ```go
  type PreviewUpdatedMsg struct {
      SessionID string
      Content   string
  }
  ```

### Selection Change

As a user, when I select a different agent from the side menu, the preview immediately updates to show that agent's output.

#### Technical / Architecture

- On selection change (up/down navigation):
  1. Immediately capture new agent's output (don't wait for next poll)
  2. Update preview content
  3. Continue normal polling from there

- Implementation:
  ```go
  type AgentSelectedMsg struct {
      Agent *Agent
  }

  func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
      switch msg := msg.(type) {
      case AgentSelectedMsg:
          // Immediate capture
          return m, m.capturePreview(msg.Agent.ID)
      }
  }
  ```

- Only one agent's output is ever captured at a time

### Pane Sizing

As a user, the preview fits within the content area without stretching or breaking the dashboard layout.

#### Technical / Architecture

- Dashboard layout is fixed relative to window size:
  ```
  ┌─────────────┬─────────────────────────────┐
  │             │                             │
  │  Side Menu  │      Content Area           │
  │   (25%)     │        (75%)                │
  │             │                             │
  │             │                             │
  ├─────────────┴─────────────────────────────┤
  │           Quick Commands (3 lines)        │
  └───────────────────────────────────────────┘
  ```

- Content area knows its dimensions from `SetSize(width, height)`

- Calculate lines to capture:
  ```go
  func (m *ContentAreaModel) availableLines() int {
      // Subtract border/padding
      return m.height - 2
  }
  ```

- Preview is truncated to fit:
  - Capture N lines where N = available lines
  - If output is shorter, show what exists
  - No scrolling - always shows the "tail" (most recent output)

- Content never overflows - the tmux capture handles line limiting

### Empty State

As a user, when no agents exist, I see a branded welcome message centered in the content area.

#### Technical / Architecture

- Layout (vertically distributed in pane):
  ```
  ┌─────────────────────────────────────────────────────────────┐
  │                                                             │
  │        Using Artificial Intelligence for coding?            │  <- top, centered
  │                      You must be                            │
  │                                                             │
  │                                                             │
  │                      crAIzy                                 │  <- middle, ASCII logo
  │                   (in ASCII art)                            │
  │                                                             │
  │                                                             │
  │                        v0.1.0                               │  <- bottom, version
  └─────────────────────────────────────────────────────────────┘
  ```

- Logo requirements:
  - Spells "crAIzy" in ASCII art
  - Case distinction maintained: lowercase `cr`, uppercase `AI`, lowercase `zy`
  - The "AI" should visually stand out (larger/bolder if possible)
  - Final design to be crafted during implementation (figlet, toilet, or hand-drawn)

- Implementation:
  ```go
  const logo = `<ASCII art for "crAIzy" - TBD>`

  func (m *ContentAreaModel) renderEmptyState() string {
      // Top text - centered
      topText := lipgloss.PlaceHorizontal(m.width, lipgloss.Center,
          "Using Artificial Intelligence for coding?\nYou must be")

      // Logo - centered
      logoBlock := lipgloss.PlaceHorizontal(m.width, lipgloss.Center, logo)

      // Version - centered at bottom
      version := lipgloss.PlaceHorizontal(m.width, lipgloss.Center, "v0.1.0")

      // Compose vertically with spacing
      content := lipgloss.JoinVertical(lipgloss.Center,
          topText,
          "",
          logoBlock,
          "",
          version,
      )

      // Center the whole block in the pane
      return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
  }
  ```

- Version should come from build-time variable or config

- Show empty state when:
  - No agents in `AgentStore`
  - After killing last agent

### Pause During Port

As a user, preview polling stops when I port into an agent and resumes when I detach.

#### Technical / Architecture

- Track porting state in model:
  ```go
  type Model struct {
      // ...
      isPortedIn bool
  }
  ```

- On port in (`tea.ExecProcess` starts):
  - Set `isPortedIn = true`
  - Polling tick handler checks this flag and skips capture

- On detach (`AgentDetachedMsg` received):
  - Set `isPortedIn = false`
  - Immediately capture preview
  - Resume polling

- Implementation:
  ```go
  func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
      switch msg := msg.(type) {
      case PreviewTickMsg:
          if m.isPortedIn {
              // Skip capture, but keep ticking
              return m, m.pollPreview()
          }
          return m, m.capturePreview(m.selectedAgentID())
      case AgentDetachedMsg:
          m.isPortedIn = false
          return m, tea.Batch(m.capturePreview(m.selectedAgentID()), m.pollPreview())
      }
  }
  ```

## Open Questions

None - all resolved.

## Out of Scope

- Scrolling through preview history
- Capturing multiple agents simultaneously
- Preview of agent that isn't selected
- Interactive preview (sending input)
