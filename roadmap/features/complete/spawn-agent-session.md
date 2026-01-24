Epic: MVP

# Spawn Agent Session

Dependencies: none

## Description

Enable spawning, listing, selecting, attaching to, and killing AI agent sessions via tmux. This is the foundational feature that allows multiple concurrent agents to be managed from the craizy dashboard.

The dashboard acts as the host/control panel. Agents run in background tmux sessions. Users can "port into" an agent (attach), work with it, then detach back to the dashboard.

## Stories

### New Agent

As a user, when I complete the 'new agent' flow, a tmux session starts with my chosen agent and name.

#### Technical / Architecture

- Session ID pattern: `craizy-{project}-{agent}-{name}`
  - `{project}`: Parent folder name where craizy was launched
  - `{agent}`: Agent type from AGENTS.yml (lowercase, sanitized)
  - `{name}`: User-entered name (sanitized: no `.` or `:`, lowercase)
  - Example: `craizy-myapp-claude-research`

- Tmux sessions are created detached (`tmux new-session -d`)

- Sessions start in the working directory where craizy was launched

- Duplicate names are rejected with an error modal

- Architecture follows service layer pattern with dependency injection:

  ```go
  // Domain model
  type Agent struct {
      ID        string    // The tmux session ID
      Project   string
      AgentType string
      Name      string
      Command   string
      WorkDir   string
      CreatedAt time.Time
  }

  // Service layer - orchestrates domain operations
  type AgentService struct {
      tmux  ITmuxClient
      store IAgentStore
  }

  func (s *AgentService) Create(project, agentType, name, command, workDir string) (*Agent, error)
  func (s *AgentService) Kill(sessionID string) error
  func (s *AgentService) List() []*Agent
  func (s *AgentService) Attach(sessionID string) tea.Cmd
  func (s *AgentService) Exists(sessionID string) bool
  func (s *AgentService) CleanupZombies(project string) error

  // Domain interfaces - implementations injected
  type ITmuxClient interface {
      CreateSession(id, command, workDir string) error
      KillSession(id string) error
      ListSessions() ([]string, error)
      AttachCmd(id string) *exec.Cmd
      SessionExists(id string) bool
  }

  type IAgentStore interface {
      Add(agent *Agent) error
      Remove(id string) error
      List() []*Agent
      Get(id string) *Agent
      Exists(id string) bool
  }
  ```

- On creation, service orchestrates:
  ```go
  func (s *AgentService) Create(...) (*Agent, error) {
      // 1. Validate uniqueness
      if s.store.Exists(sessionID) {
          return nil, ErrDuplicateName
      }
      // 2. Create tmux session
      if err := s.tmux.CreateSession(sessionID, command, workDir); err != nil {
          return nil, err
      }
      // 3. Store in memory
      agent := &Agent{...}
      s.store.Add(agent)
      // 4. Return (caller sends AgentsUpdatedMsg)
      return agent, nil
  }
  ```

- Errors (tmux failure, duplicate name) shown in modal

### Agent List

As a user, I can see currently active agents listed in the side menu, and navigate between them with up/down keys.

#### Technical / Architecture

- Side menu displays agents from `AgentStore`

- List is "selectable" using Bubble Tea's `list` or `bubbles/list` component

- Single agent: auto-selected

- Multiple agents: up/down navigates, selection highlighted

- State updates via message:
  ```go
  type AgentsUpdatedMsg struct {
      Agents []*Agent
  }
  ```

- `AgentService.List()` is the source of truth (backed by `IAgentStore`)

- In-memory implementation for MVP:
  ```go
  type MemoryAgentStore struct {
      agents map[string]*Agent
      mu     sync.RWMutex
  }
  // Implements IAgentStore
  ```

- Side menu subscribes to `AgentsUpdatedMsg` and re-renders

- Future updates (status changes, git activity) will also trigger `AgentsUpdatedMsg`

### Port to Agent

As a user, when I press Enter on the dashboard, I attach to the currently selected agent's tmux session. When I detach (Ctrl+B, D), I return to the dashboard.

#### Technical / Architecture

- "Porting" means:
  1. Suspend the Bubble Tea TUI
  2. Execute `tmux attach -t {sessionID}` (takes over terminal)
  3. User interacts with agent
  4. User detaches with Ctrl+B, D
  5. `tmux attach` exits, TUI resumes

- Service provides attach command:
  ```go
  func (s *AgentService) Attach(sessionID string) tea.Cmd {
      cmd := s.tmux.AttachCmd(sessionID)
      return tea.ExecProcess(cmd, func(err error) tea.Msg {
          return AgentDetachedMsg{SessionID: sessionID, Err: err}
      })
  }
  ```

- Quick commands bar shows hint: `enter - port to agent`

- When inside agent, tmux status bar shows: `Detach: Ctrl+B, D`

### Kill Agent

As a user, I can press `k` to kill the currently selected agent session.

#### Technical / Architecture

- New quick command: `k - kill agent`

- On `k` keypress:
  1. Get currently selected agent from side menu
  2. Call `AgentService.Kill(sessionID)`
  3. Caller sends `AgentsUpdatedMsg` to update UI

- Service orchestrates kill:
  ```go
  func (s *AgentService) Kill(sessionID string) error {
      // 1. Kill tmux session
      if err := s.tmux.KillSession(sessionID); err != nil {
          return err
      }
      // 2. Remove from store
      s.store.Remove(sessionID)
      return nil
  }
  ```

- If no agent selected, `k` does nothing (or shows "no agent selected")

### Fresh Start Protocol

As a user, when the application starts, any zombie agent sessions from a previous crash are cleaned up.

#### Technical / Architecture

- On startup, before TUI renders:
  1. Query tmux: `tmux list-sessions -F "#{session_name}"`
  2. Filter sessions matching `craizy-{project}-*`
  3. Kill all matching sessions
  4. Start with clean slate

- Service orchestrates cleanup:
  ```go
  func (s *AgentService) CleanupZombies(project string) error {
      prefix := fmt.Sprintf("craizy-%s-", project)
      sessions, _ := s.tmux.ListSessions()
      for _, sessionID := range sessions {
          if strings.HasPrefix(sessionID, prefix) {
              s.tmux.KillSession(sessionID)
          }
      }
      // Store starts empty, no cleanup needed there
      return nil
  }
  ```

- Called in `main.go` before `tea.NewProgram()`

- Future enhancement: reconcile with database, offer to recover instead of kill

## Open Questions

- Should `k` require confirmation before killing? (Leaning: no for MVP, add later if needed)
- Should the tmux status bar be customized to show agent name + detach hint? (Leaning: yes)

## Out of Scope

- Database persistence (later feature)
- Preview pane showing agent output (later feature)
- Agent status indicators (later feature)
- Director/Lead/Worker hierarchy (later feature)
- Reconciliation with recovery (later - for now, just kill zombies)
