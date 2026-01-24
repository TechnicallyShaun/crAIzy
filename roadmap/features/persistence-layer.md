Epic: MVP

# Persistence Layer

Dependencies: [spawn-agent-session.md](./spawn-agent-session.md)

## Description

Implement a persistent storage layer using hexagonal architecture with an event-driven pattern. Domain services emit events; adapters subscribe and react (database persistence, tmux operations, UI updates).

The database choice (Dolt vs SQLite) is deferred as an implementation detail behind the `IAgentStore` port interface. The architecture supports swapping implementations without changing domain logic.

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                      Application Layer                          │
│                   (TUI Commands, Use Cases)                     │
└─────────────────────────────┬───────────────────────────────────┘
                              │ calls
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                       Domain Layer                              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐  │
│  │   Models    │  │  Services   │  │    Event Dispatcher     │  │
│  │  - Agent    │  │AgentService │  │  Publish(event)         │  │
│  │  - WorkItem │  │WorkService  │  │  Subscribe(handler)     │  │
│  │  - Message  │  │             │  │                         │  │
│  └─────────────┘  └─────────────┘  └─────────────────────────┘  │
└─────────────────────────────┬───────────────────────────────────┘
                              │ emits domain events
                              ▼
        ┌─────────────────────┼─────────────────────┐
        ▼                     ▼                     ▼
   ┌─────────┐          ┌──────────┐         ┌──────────┐
   │ Store   │          │ Terminal │         │ Preview  │
   │ Port    │          │ Port     │         │ Port     │
   └────┬────┘          └────┬─────┘         └────┬─────┘
        │                    │                    │
        ▼                    ▼                    ▼
   ┌─────────┐          ┌──────────┐         ┌──────────┐
   │Dolt/    │          │  Tmux    │         │ Preview  │
   │SQLite   │          │  Client  │         │ Service  │
   └─────────┘          └──────────┘         └──────────┘
              Adapters (swappable)
```

## Stories

### Domain Events Infrastructure

As a developer, I can emit domain events from services, and adapters receive them to perform side effects.

#### Technical / Architecture

- Event dispatcher interface:
  ```go
  // Domain event base
  type Event interface {
      EventType() string
      OccurredAt() time.Time
  }

  // Dispatcher - lives in domain layer
  type IEventDispatcher interface {
      Publish(event Event)
      Subscribe(eventType string, handler EventHandler)
  }

  type EventHandler func(event Event)
  ```

- In-memory dispatcher implementation:
  ```go
  type EventDispatcher struct {
      handlers map[string][]EventHandler
      mu       sync.RWMutex
  }

  func (d *EventDispatcher) Publish(event Event) {
      d.mu.RLock()
      defer d.mu.RUnlock()
      for _, handler := range d.handlers[event.EventType()] {
          handler(event)
      }
  }

  func (d *EventDispatcher) Subscribe(eventType string, handler EventHandler) {
      d.mu.Lock()
      defer d.mu.Unlock()
      d.handlers[eventType] = append(d.handlers[eventType], handler)
  }
  ```

- Events are synchronous for MVP (handler blocks publisher)
- Future: async with channels if needed

### Agent Domain Events

As a developer, when agent lifecycle methods are called, appropriate events are emitted.

#### Technical / Architecture

- Agent events:
  ```go
  type AgentCreated struct {
      Agent     *Agent
      Timestamp time.Time
  }
  func (e AgentCreated) EventType() string { return "agent.created" }
  func (e AgentCreated) OccurredAt() time.Time { return e.Timestamp }

  type AgentKilled struct {
      AgentID   string
      Timestamp time.Time
  }
  func (e AgentKilled) EventType() string { return "agent.killed" }
  func (e AgentKilled) OccurredAt() time.Time { return e.Timestamp }

  type AgentStatusChanged struct {
      AgentID   string
      OldStatus string
      NewStatus string
      Timestamp time.Time
  }
  func (e AgentStatusChanged) EventType() string { return "agent.status_changed" }
  func (e AgentStatusChanged) OccurredAt() time.Time { return e.Timestamp }
  ```

- AgentService emits events after successful operations:
  ```go
  type AgentService struct {
      dispatcher IEventDispatcher
      store      IAgentStore
      tmux       ITmuxClient
  }

  func (s *AgentService) Create(project, agentType, name, command, workDir string) (*Agent, error) {
      // 1. Validate
      sessionID := buildSessionID(project, agentType, name)
      if s.store.Exists(sessionID) {
          return nil, ErrDuplicateName
      }

      // 2. Build domain object
      agent := &Agent{
          ID:        sessionID,
          Project:   project,
          AgentType: agentType,
          Name:      name,
          Command:   command,
          WorkDir:   workDir,
          Status:    "pending",
          CreatedAt: time.Now(),
      }

      // 3. Emit event - adapters react
      s.dispatcher.Publish(AgentCreated{Agent: agent, Timestamp: time.Now()})

      return agent, nil
  }
  ```

- Adapters subscribe and handle:
  ```go
  // In application bootstrap
  func wireAdapters(dispatcher IEventDispatcher, store IAgentStore, tmux ITmuxClient) {
      // Store adapter - persists to database
      dispatcher.Subscribe("agent.created", func(e Event) {
          evt := e.(AgentCreated)
          store.Add(evt.Agent)
      })

      // Tmux adapter - creates session
      dispatcher.Subscribe("agent.created", func(e Event) {
          evt := e.(AgentCreated)
          tmux.CreateSession(evt.Agent.ID, evt.Agent.Command, evt.Agent.WorkDir)
      })

      // Kill handlers
      dispatcher.Subscribe("agent.killed", func(e Event) {
          evt := e.(AgentKilled)
          tmux.KillSession(evt.AgentID)
          store.Remove(evt.AgentID)
      })
  }
  ```

### Database Store Implementation

As a developer, I can persist agents to a database that survives restarts.

#### Technical / Architecture

- `IAgentStore` interface defined in [spawn-agent-session.md](./spawn-agent-session.md)

- Database-backed implementation (Dolt or SQLite):
  ```go
  type DBAgentStore struct {
      db *sql.DB
  }

  func (s *DBAgentStore) Add(agent *Agent) error {
      _, err := s.db.Exec(`
          INSERT INTO agents (id, project, agent_type, name, command, work_dir, status, created_at)
          VALUES (?, ?, ?, ?, ?, ?, ?, ?)
      `, agent.ID, agent.Project, agent.AgentType, agent.Name, agent.Command, agent.WorkDir, agent.Status, agent.CreatedAt)
      return err
  }

  func (s *DBAgentStore) List() []*Agent {
      rows, _ := s.db.Query(`SELECT id, project, agent_type, name, command, work_dir, status, created_at FROM agents`)
      defer rows.Close()
      var agents []*Agent
      for rows.Next() {
          a := &Agent{}
          rows.Scan(&a.ID, &a.Project, &a.AgentType, &a.Name, &a.Command, &a.WorkDir, &a.Status, &a.CreatedAt)
          agents = append(agents, a)
      }
      return agents
  }

  // ... other methods
  ```

- Schema:
  ```sql
  CREATE TABLE agents (
      id TEXT PRIMARY KEY,
      project TEXT NOT NULL,
      agent_type TEXT NOT NULL,
      name TEXT NOT NULL,
      command TEXT NOT NULL,
      work_dir TEXT NOT NULL,
      status TEXT DEFAULT 'active',
      created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
      terminated_at DATETIME
  );
  ```

- Q: Dolt vs SQLite - deferred decision, interface is the same

### Database Migration

As a developer, when the application starts, the database schema is created if it doesn't exist.

#### Technical / Architecture

- Embedded migrations in binary:
  ```go
  //go:embed migrations/*.sql
  var migrations embed.FS

  func Migrate(db *sql.DB) error {
      // Read and execute migrations in order
      entries, _ := migrations.ReadDir("migrations")
      for _, entry := range entries {
          content, _ := migrations.ReadFile("migrations/" + entry.Name())
          _, err := db.Exec(string(content))
          if err != nil {
              return fmt.Errorf("migration %s failed: %w", entry.Name(), err)
          }
      }
      return nil
  }
  ```

- Migration files:
  ```
  internal/store/migrations/
  ├── 001_create_agents.sql
  ├── 002_create_work_items.sql
  └── 003_create_messages.sql
  ```

- Called at startup before TUI:
  ```go
  func main() {
      db := openDatabase()
      store.Migrate(db)
      // ... continue startup
  }
  ```

- For MVP: simple sequential execution
- Future: migration versioning table to track applied migrations

### Startup Reconciliation

As a user, when the application starts, database state is reconciled with actual tmux sessions.

#### Technical / Architecture

- On startup:
  1. Load agents from database
  2. List actual tmux sessions
  3. Reconcile:
     - DB has agent, tmux missing → mark as terminated (crashed)
     - Tmux has session, DB missing → orphan, kill it (zombie from before DB)
     - Both match → healthy, keep

- Reconciliation logic:
  ```go
  func (s *AgentService) Reconcile() error {
      dbAgents := s.store.List()
      tmuxSessions, _ := s.tmux.ListSessions()
      tmuxSet := toSet(tmuxSessions)

      for _, agent := range dbAgents {
          if !tmuxSet[agent.ID] {
              // Agent in DB but tmux session gone - mark terminated
              s.store.UpdateStatus(agent.ID, "terminated")
          }
      }

      prefix := fmt.Sprintf("craizy-%s-", s.project)
      for _, session := range tmuxSessions {
          if strings.HasPrefix(session, prefix) {
              if s.store.Get(session) == nil {
                  // Tmux session exists but not in DB - zombie, kill it
                  s.tmux.KillSession(session)
              }
          }
      }
      return nil
  }
  ```

- Replaces simple "kill all zombies" from spawn-agent-session with smarter reconciliation

## Domain Models (Future Stories)

These models are documented for context but implemented in later features:

### WorkItem
```go
type WorkItem struct {
    ID          string
    ParentID    *string     // NULL = top-level
    ItemType    string      // epic, feature, story, task
    Title       string
    Description string
    Status      string      // pending, in_progress, blocked, done
    AssignedTo  *string     // Agent ID
    CreatedBy   *string     // Agent ID
    DependsOn   []string    // WorkItem IDs
    CreatedAt   time.Time
    CompletedAt *time.Time
}
```

### Message

See [messaging.md](./messaging.md) for full Message model definition.

## Open Questions

- Dolt vs SQLite: Leaning Dolt for built-in audit trail and time-travel queries. Decision can be deferred since both implement same interface.
- Should events be async (channel-based) or sync? Starting sync for simplicity.
- Migration tooling: embed.FS simple approach vs golang-migrate for rollbacks?
- Should reconciliation prompt user before killing orphan sessions?

## Out of Scope

- WorkItem persistence (later feature - hierarchy/orchestration)
- Message persistence (later feature - agent communication)
- Audit log table (Dolt provides this implicitly; explicit table if using SQLite)
- Remote database / multi-machine sync
- Backup/restore utilities
