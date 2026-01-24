# Architecture Deep Dive

## Core Principle

The system is **generic and recursive** - the same patterns apply at every level of hierarchy, regardless of depth. This makes it applicable to software development, research projects, physical projects, or any domain with decomposable work.

---

## Data Layer: SQLite as Central Nervous System

Everything lives in the database. The database is the **single source of truth**.

### Schema Overview

```sql
-- Agent Pool (available for hire)
CREATE TABLE agent_types (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    command TEXT NOT NULL,
    capabilities TEXT,  -- JSON array of what this agent can do
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Active Agents (currently hired into the system)
CREATE TABLE agents (
    id TEXT PRIMARY KEY,
    agent_type_id TEXT REFERENCES agent_types(id),
    role TEXT NOT NULL,  -- 'director', 'lead', 'worker'
    tmux_session TEXT NOT NULL,
    parent_id TEXT REFERENCES agents(id),  -- Who hired me?
    status TEXT DEFAULT 'active',  -- 'active', 'idle', 'terminated'
    hired_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    terminated_at DATETIME
);

-- Work Breakdown Structure (recursive)
CREATE TABLE work_items (
    id TEXT PRIMARY KEY,
    parent_id TEXT REFERENCES work_items(id),  -- NULL = top-level epic
    item_type TEXT NOT NULL,  -- 'epic', 'feature', 'story', 'task'
    title TEXT NOT NULL,
    description TEXT,
    status TEXT DEFAULT 'pending',  -- 'pending', 'in_progress', 'blocked', 'done'
    assigned_to TEXT REFERENCES agents(id),
    created_by TEXT REFERENCES agents(id),
    depends_on TEXT,  -- JSON array of work_item IDs
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME
);

-- Message Queue
CREATE TABLE messages (
    id TEXT PRIMARY KEY,
    from_agent TEXT REFERENCES agents(id),
    to_agent TEXT REFERENCES agents(id),
    work_item_id TEXT REFERENCES work_items(id),  -- Context
    message_type TEXT NOT NULL,  -- 'task_assignment', 'question', 'completion', 'status_update'
    content TEXT NOT NULL,
    read BOOLEAN DEFAULT FALSE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    read_at DATETIME
);

-- Audit Log (for learning/improvement)
CREATE TABLE audit_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    agent_id TEXT REFERENCES agents(id),
    action TEXT NOT NULL,  -- 'hire', 'fire', 'assign', 'complete', 'message', 'tool_call'
    details TEXT,  -- JSON blob
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

---

## Hierarchy: Org Chart in a Database

```
Director (singular, always exists)
    │
    ├── Lead (Feature: Auth System)
    │       ├── Worker (Task: Create login form)
    │       ├── Worker (Task: Set up JWT)
    │       └── Worker (Task: Write auth tests)
    │
    └── Lead (Feature: Dashboard)
            ├── Worker (Task: Design layout)
            └── Worker (Task: Implement charts)
```

### Key Properties

1. **Director is singular** - One per system instance
2. **Leads own features** - One lead per feature/major chunk
3. **Workers are ephemeral** - Spin up, complete task, terminate
4. **Parent-child via `parent_id`** - Every agent knows who hired them
5. **Any agent can query the structure** - Via tool calls to the database

---

## Agent Capabilities: Tool-Based Actions

Agents don't act freely - they **call tools** which the Go harness executes. This creates structure and auditability.

### Director Tools

```go
// director can call these
tools := map[string]Tool{
    "hire_lead": HireLeadTool{},       // Spawn a lead for a feature
    "view_structure": ViewOrgTool{},    // See all active agents
    "view_workload": ViewWorkTool{},    // See all epics/features
    "send_message": SendMessageTool{},  // Message a lead
    "terminate": TerminateTool{},       // Fire an agent
}
```

### Lead Tools

```go
tools := map[string]Tool{
    "breakdown_feature": BreakdownTool{},  // Decompose into stories/tasks
    "hire_worker": HireWorkerTool{},       // Spawn a worker for a task
    "view_team": ViewTeamTool{},           // See my workers
    "view_tasks": ViewTasksTool{},         // See my assigned work items
    "mark_complete": MarkCompleteTool{},   // Roll up completion
    "send_message": SendMessageTool{},     // Message director or workers
    "ask_human": AskHumanTool{},           // Escalate question to human
}
```

### Worker Tools

```go
tools := map[string]Tool{
    "view_task": ViewTaskTool{},           // See my assigned task
    "mark_done": MarkDoneTool{},           // Signal completion
    "ask_lead": AskLeadTool{},             // Question for my lead
    "request_help": RequestHelpTool{},     // Need another worker
}
```

---

## The Hire Flow

When any agent calls `hire_*`:

```
Agent calls: hire_worker(task_id: "123", agent_type: "claude")
                          │
                          ▼
              ┌─────────────────────┐
              │   Go Harness        │
              │                     │
              │ 1. Query agent_types│
              │ 2. Create tmux      │
              │ 3. Insert agents row│
              │ 4. Link to parent   │
              │ 5. Assign work_item │
              │ 6. Send init message│
              │ 7. Log to audit     │
              └─────────────────────┘
                          │
                          ▼
              New worker session with context:
              "You are a worker. Your task is X.
               Your lead is Y. Call mark_done when complete."
```

---

## The Completion Roll-Up

When a worker completes:

```
Worker calls: mark_done(summary: "Implemented login form")
                          │
                          ▼
              ┌─────────────────────┐
              │   Go Harness        │
              │                     │
              │ 1. Update work_item │
              │    status = 'done'  │
              │ 2. Update agent     │
              │    status = 'term'  │
              │ 3. Kill tmux session│
              │ 4. Message parent   │
              │    (the lead)       │
              │ 5. Check: all tasks │
              │    for this feature │
              │    done?            │
              │ 6. If yes, notify   │
              │    lead to roll up  │
              │ 7. Log to audit     │
              └─────────────────────┘
```

The same pattern applies when a Lead completes a feature → notifies Director.

---

## Message Flow

Messages are **stored in DB** and **injected into tmux**.

```go
func SendMessage(from, to, content string, workItemID string) {
    // 1. Insert into messages table
    db.Exec(`INSERT INTO messages (id, from_agent, to_agent, content, work_item_id)
             VALUES (?, ?, ?, ?, ?)`, uuid(), from, to, content, workItemID)

    // 2. Get recipient's tmux session
    var tmuxSession string
    db.QueryRow(`SELECT tmux_session FROM agents WHERE id = ?`, to).Scan(&tmuxSession)

    // 3. Inject into their session
    exec.Command("tmux", "send-keys", "-t", tmuxSession,
        fmt.Sprintf("\n[MESSAGE from %s]: %s\n", from, content), "Enter").Run()

    // 4. Log
    logAudit(from, "message", map[string]string{"to": to, "content": content})
}
```

---

## Work Breakdown Structure

Recursive by nature. An epic contains features, features contain stories, stories contain tasks.

```
Epic: "Build Authentication System"
├── Feature: "User Registration"
│   ├── Story: "Email signup flow"
│   │   ├── Task: "Create registration form"
│   │   ├── Task: "Email validation"
│   │   └── Task: "Welcome email"
│   └── Story: "Social auth"
│       ├── Task: "Google OAuth"
│       └── Task: "GitHub OAuth"
└── Feature: "Login/Logout"
    └── ...
```

The `item_type` field is just a label - the structure is enforced by `parent_id`. You could go deeper (sub-tasks) or shallower depending on the project.

---

## Agent's View of the World

Each agent gets a **context window** showing what they need to know:

### Director sees:

```
SYSTEM STATUS
─────────────
Active Leads: 3
Active Workers: 7

CURRENT EPICS
─────────────
[IP] Build Auth System (2/5 features done)
[IP] Dashboard Redesign (0/3 features done)
[PENDING] API v2 Migration

MY LEADS
────────
Lead-001: Alice (Auth System) - 3 workers
Lead-002: Bob (Dashboard) - 2 workers
Lead-003: Carol (API v2) - idle
```

### Lead sees:

```
MY FEATURE: User Registration
─────────────────────────────
Director: Director-001
Status: In Progress

STORIES
───────
[DONE] Email signup flow (3/3 tasks)
[IP] Social auth (1/2 tasks)

MY WORKERS
──────────
Worker-005: Google OAuth (in progress)
Worker-006: GitHub OAuth (in progress)
```

### Worker sees:

```
MY TASK: Implement Google OAuth
─────────────────────────────────
Lead: Lead-001 (Alice)
Feature: User Registration
Story: Social auth

DESCRIPTION
───────────
Implement Google OAuth login using the oauth2 library.
Redirect to /auth/google/callback.
Store tokens in session.

DEPENDS ON
──────────
(none - ready to start)
```

---

## The Generic Loop

Every agent, regardless of role, follows the same loop:

```
1. Wake up with context (who am I, what's my work)
2. Query: What's my current state? (view tools)
3. Decide: What action to take?
4. Execute: Call a tool
5. Observe: Tool returns result
6. Repeat until: work complete
7. Call completion tool → terminate
```

This is **recursive and role-agnostic**. The only difference is which tools are available.

---

## Audit Trail for Learning

Every action is logged:

```sql
SELECT * FROM audit_log WHERE agent_id = 'worker-005' ORDER BY created_at;

-- Results:
-- | action      | details                                           |
-- |-------------|---------------------------------------------------|
-- | hire        | {"by": "lead-001", "for_task": "task-123"}        |
-- | view_task   | {}                                                |
-- | tool_call   | {"tool": "bash", "cmd": "npm install oauth2"}     |
-- | tool_call   | {"tool": "edit", "file": "auth.js"}               |
-- | mark_done   | {"summary": "Implemented Google OAuth"}           |
```

This creates a rich dataset for:
- Analyzing common patterns
- Identifying bottlenecks
- Training future automations
- Debugging failures

---

## Open Questions

1. **How deep can hierarchy go?** Director → Lead → Worker is 3 levels. Do we need Lead → Sub-Lead → Worker? The schema supports it.

2. **Concurrency limits?** How many workers can one lead manage? Should there be guardrails?

3. **Human in the loop?** Where exactly does the human approve/intervene? Just at epic creation? Or also at feature breakdown?

4. **Agent type specialization?** Should certain agent types only be hirable for certain roles? (e.g., Claude for leads, Gemini for workers)

5. **Cross-feature dependencies?** What if Task A in Feature 1 depends on Task B in Feature 2?

---

## Next Steps

- [ ] Prototype the SQLite schema
- [ ] Build the tool interfaces in Go
- [ ] Create a simple Director → Lead → Worker flow
- [ ] Test with one epic, one feature, two tasks
- [ ] Add the message injection to tmux
- [ ] Build the context view for each role
