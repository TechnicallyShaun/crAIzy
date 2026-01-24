Epic: MVP

# Messaging

Dependencies: [spawn-agent-session.md](./spawn-agent-session.md), [persistence-layer.md](./persistence-layer.md)

## Description

A messaging system that allows agents (and humans) to communicate with each other. Messages are stored in the database and delivered to recipients via tmux injection (if active) or queued for delivery on startup (if inactive).

The architecture is **domain-layer first**:
- All business logic lives in the domain layer (`MessageService`)
- CLI commands are thin wrappers that parse args and call domain methods
- In-process side effects (tmux notification, state updates) happen in the domain layer
- The database is the shared state - no daemon required

"Human" is a valid participant, treated the same as any agent at the data layer.

## Stories

### Send Message

As an agent, I can send a message to another agent (or human) using the CLI.

#### Technical / Architecture

- CLI command:
  ```bash
  craizy msg send --from <sender-id> --to <recipient-id> --type <type> --content "..."

  # Examples:
  craizy msg send --from worker-001 --to lead-001 --type question --content "Which auth library?"
  craizy msg send --from lead-001 --to human --type decision --content "OAuth or JWT?"
  ```

- Message types:
  - `question` - Needs answer before proceeding
  - `answer` - Response to a question
  - `assignment` - Work being delegated
  - `completion` - Task/work finished
  - `status` - Progress update
  - `info` - General information

- Domain layer:
  ```go
  type Message struct {
      ID          string
      From        string     // agent ID or "human"
      To          string     // agent ID or "human"
      Type        string     // question, answer, assignment, completion, status, info
      Content     string
      WorkItemID  *string    // optional context reference
      Read        bool
      CreatedAt   time.Time
      ReadAt      *time.Time
  }

  type IMessageStore interface {
      Save(msg *Message) error
      MarkRead(id string) error
      ListUnread(recipientID string) ([]*Message, error)
      List(recipientID string, limit int) ([]*Message, error)
      Get(id string) (*Message, error)
  }

  type MessageService struct {
      store IMessageStore
      tmux  ITmuxClient
      agents IAgentStore
  }

  func (s *MessageService) Send(from, to, msgType, content string, workItemID *string) (*Message, error) {
      msg := &Message{
          ID:         uuid(),
          From:       from,
          To:         to,
          Type:       msgType,
          Content:    content,
          WorkItemID: workItemID,
          Read:       false,
          CreatedAt:  time.Now(),
      }

      // 1. Persist to DB
      if err := s.store.Save(msg); err != nil {
          return nil, err
      }

      // 2. If recipient is active, deliver immediately
      if s.isActive(to) {
          s.deliverToTmux(msg)
          s.store.MarkRead(msg.ID)
      }

      return msg, nil
  }

  func (s *MessageService) isActive(agentID string) bool {
      if agentID == "human" {
          return false  // Human messages stay unread until explicitly viewed
      }
      agent := s.agents.Get(agentID)
      return agent != nil && s.tmux.SessionExists(agent.TmuxSession)
  }

  func (s *MessageService) deliverToTmux(msg *Message) {
      agent := s.agents.Get(msg.To)
      if agent == nil {
          return
      }

      notification := fmt.Sprintf("\n[MESSAGE from %s (%s)]: %s\n",
          msg.From, msg.Type, msg.Content)

      s.tmux.SendKeys(agent.TmuxSession, notification)
  }
  ```

- CLI is a thin wrapper:
  ```go
  // CLI just parses and calls domain
  func cmdMessageSend(cmd *cobra.Command, args []string) {
      from, _ := cmd.Flags().GetString("from")
      to, _ := cmd.Flags().GetString("to")
      msgType, _ := cmd.Flags().GetString("type")
      content, _ := cmd.Flags().GetString("content")

      // Get domain service (initialized at app startup)
      svc := domain.GetMessageService()

      // All logic happens in domain layer
      msg, err := svc.Send(from, to, msgType, content, nil)
      if err != nil {
          fmt.Fprintf(os.Stderr, "Error: %v\n", err)
          os.Exit(1)
      }

      fmt.Printf("Message sent: %s\n", msg.ID)
  }
  ```

- Domain service initialization (shared across CLI and TUI):
  ```go
  // internal/domain/init.go
  var (
      messageService *MessageService
      agentService   *AgentService
      once           sync.Once
  )

  func Init(db *sql.DB) {
      once.Do(func() {
          tmux := infra.NewTmuxClient()
          agentStore := infra.NewSQLiteAgentStore(db)
          messageStore := infra.NewSQLiteMessageStore(db)

          agentService = NewAgentService(tmux, agentStore)
          messageService = NewMessageService(messageStore, tmux, agentStore)
      })
  }

  func GetMessageService() *MessageService { return messageService }
  func GetAgentService() *AgentService { return agentService }
  ```

### List Messages

As an agent (or human), I can list my messages via CLI.

#### Technical / Architecture

- CLI commands:
  ```bash
  craizy msg list --for <agent-id>           # All messages
  craizy msg list --for <agent-id> --unread  # Only unread
  craizy msg list --for human --unread       # Human's inbox

  # Short form
  craizy msg ls --for worker-001
  ```

- Output format:
  ```
  ID          FROM        TYPE        TIME                 CONTENT
  msg-001     lead-001    assignment  2024-01-24 10:30:00  Implement OAuth...
  msg-002     worker-002  question    2024-01-24 10:45:00  Which library...

  2 messages (1 unread)
  ```

- Domain layer:
  ```go
  func (s *MessageService) ListUnread(recipientID string) ([]*Message, error) {
      return s.store.ListUnread(recipientID)
  }

  func (s *MessageService) List(recipientID string, limit int) ([]*Message, error) {
      return s.store.List(recipientID, limit)
  }
  ```

### Read Message

As an agent (or human), I can read a specific message and mark it as read.

#### Technical / Architecture

- CLI command:
  ```bash
  craizy msg read <message-id>
  ```

- Output:
  ```
  From:    lead-001
  To:      worker-001
  Type:    assignment
  Time:    2024-01-24 10:30:00
  Context: work-item-123

  Content:
  ─────────────────────────────────
  Implement Google OAuth login using the oauth2 library.
  Redirect to /auth/google/callback.
  Store tokens in session.
  ─────────────────────────────────

  [Marked as read]
  ```

- Domain layer:
  ```go
  func (s *MessageService) Read(messageID string) (*Message, error) {
      msg, err := s.store.Get(messageID)
      if err != nil {
          return nil, err
      }

      if !msg.Read {
          s.store.MarkRead(messageID)
      }

      return msg, nil
  }
  ```

### Startup Delivery

As an agent, when I start up, I receive all unread messages that were sent while I was inactive.

#### Technical / Architecture

- On agent creation (in `AgentService.Create`):
  ```go
  func (s *AgentService) Create(...) (*Agent, error) {
      // ... create tmux session ...

      // Deliver any queued messages
      s.deliverQueuedMessages(agent)

      return agent, nil
  }

  func (s *AgentService) deliverQueuedMessages(agent *Agent) {
      messages, _ := s.messageSvc.ListUnread(agent.ID)

      if len(messages) == 0 {
          return
      }

      // Header
      s.tmux.SendKeys(agent.TmuxSession,
          fmt.Sprintf("\n=== %d queued messages ===\n", len(messages)))

      // Deliver each
      for _, msg := range messages {
          notification := fmt.Sprintf("[%s from %s]: %s\n",
              msg.Type, msg.From, msg.Content)
          s.tmux.SendKeys(agent.TmuxSession, notification)
          s.messageSvc.MarkRead(msg.ID)
      }

      s.tmux.SendKeys(agent.TmuxSession, "=== End of queued messages ===\n\n")
  }
  ```

### Unread Count

As a participant, I can check how many unread messages I have.

#### Technical / Architecture

- CLI command:
  ```bash
  craizy msg count --for <agent-id>

  # Output:
  3 unread messages
  ```

- Useful for:
  - Scripts/hooks checking if there's work
  - TUI polling for notification badge
  - Director checking human's inbox

### Human as Participant

As a human, I am a valid recipient in the messaging system, with the same capabilities as agents.

#### Technical / Architecture

- "human" is a reserved agent ID
- Messages to human are never auto-delivered (no tmux session)
- Human reads messages via:
  - CLI: `craizy msg list --for human`
  - Director: "Read my messages" (Director queries and summarizes)
  - Future: TUI inbox section
  - Future: Web UI

- Human sends messages via:
  - CLI: `craizy msg send --from human --to director --content "..."`
  - Future: TUI compose
  - Future: Web UI

- No special handling at data layer - human is just an ID

## Package Structure

```
internal/
├── domain/                    # Business logic (the core)
│   ├── init.go                # Service initialization
│   ├── message_service.go     # MessageService
│   ├── agent_service.go       # AgentService
│   ├── message.go             # Message entity
│   └── interfaces.go          # IMessageStore, ITmuxClient, etc.
│
├── infra/                     # Infrastructure implementations
│   ├── sqlite_message_store.go
│   ├── sqlite_agent_store.go
│   ├── memory_agent_store.go  # MVP fallback
│   └── tmux_client.go
│
├── cli/                       # Thin CLI wrappers
│   ├── msg_send.go
│   ├── msg_list.go
│   └── msg_read.go
│
└── tui/                       # Dashboard (also calls domain)
    └── ...
```

Flow:
```
CLI/TUI ──► domain.MessageService.Send() ──► IMessageStore.Save()
                      │                           │
                      │                           ▼
                      │                      SQLite/Memory
                      │
                      └──► ITmuxClient.SendKeys() (in-process side effect)
```

## Database Schema

```sql
CREATE TABLE messages (
    id TEXT PRIMARY KEY,
    from_agent TEXT NOT NULL,          -- sender ID or "human"
    to_agent TEXT NOT NULL,            -- recipient ID or "human"
    type TEXT NOT NULL,                -- question, answer, assignment, etc.
    content TEXT NOT NULL,
    work_item_id TEXT,                 -- optional FK to work_items
    read BOOLEAN DEFAULT FALSE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    read_at DATETIME
);

CREATE INDEX idx_messages_to_unread ON messages(to_agent, read);
CREATE INDEX idx_messages_to_agent ON messages(to_agent, created_at);
```

## CLI Command Summary

| Command | Description |
|---------|-------------|
| `craizy msg send --from X --to Y --type T --content "..."` | Send a message |
| `craizy msg list --for X [--unread]` | List messages for recipient |
| `craizy msg read <id>` | Read a message, mark as read |
| `craizy msg count --for X` | Count unread messages |

## Open Questions

None - all resolved.

## Out of Scope

- Reply threading (messages are flat for now)
- Attachments
- TUI inbox UI (future feature)
- Web UI (future feature)
- Message deletion/archival
- Delivery receipts beyond read/unread
