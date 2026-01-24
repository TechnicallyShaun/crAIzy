# Process Model: Generic Recursive Decomposition

## Core Insight

The system does **one thing** at every level:

> Take a body of work → Break it into smaller bodies of work → Delegate or execute

This is the same whether you're:
- Turning a **vision** into **features** (planning)
- Turning a **feature** into **tasks** (implementation)
- Turning a **question** into **sub-questions** (research)

The hierarchy depth and labels are just context. The process is identical.

---

## Human Interface

The human provides **user stories** or **intent**:

```
As a user, I want to authenticate via Google OAuth
so that I don't need to remember another password.

Acceptance:
- [ ] "Login with Google" button on login page
- [ ] Redirects to Google, returns with token
- [ ] Creates account if first login
- [ ] Links to existing account if email matches
```

This is the **input**. Everything else is autonomous.

Human can also:
- Answer questions when escalated
- Approve major decisions (optional gate)
- Monitor progress via dashboard

---

## The Universal Loop

Every agent, regardless of role or phase, runs:

```
┌─────────────────────────────────────────────────────┐
│                                                     │
│   1. RECEIVE work item (from parent or human)       │
│                         │                           │
│                         ▼                           │
│   2. ANALYZE - Can I do this directly?              │
│         │                                           │
│         ├── YES → Execute, mark done, notify parent │
│         │                                           │
│         └── NO → Decompose into smaller items       │
│                         │                           │
│                         ▼                           │
│   3. QUESTIONS? ────YES──→ Send to parent, WAIT     │
│         │                                           │
│         NO                                          │
│         │                                           │
│         ▼                                           │
│   4. DELEGATE - Hire children for each sub-item     │
│                         │                           │
│                         ▼                           │
│   5. MONITOR - Wait for children to complete        │
│                         │                           │
│                         ▼                           │
│   6. ROLL UP - All done? Mark self done, terminate  │
│                                                     │
└─────────────────────────────────────────────────────┘
```

This loop is **the same** for Director, Lead, Worker. The difference is just:
- **Worker**: Usually executes directly (step 2 → YES)
- **Lead**: Usually decomposes (step 2 → NO)
- **Director**: Always decomposes

---

## Work Item Types

Rather than rigid types, work items have a **nature** that guides handling:

| Nature | Description | Typical Action |
|--------|-------------|----------------|
| `vision` | High-level goal, abstract | Decompose into features |
| `feature` | User-facing capability | Decompose into stories/tasks |
| `story` | Specific user scenario | Decompose into tasks |
| `task` | Atomic unit of work | Execute directly |
| `question` | Needs answer before proceeding | Research or escalate |
| `decision` | Needs choice between options | Escalate to human or parent |
| `definition` | Needs clarification | Research or ask |

The agent **decides** what to do based on the nature + their role + the content.

---

## Questions Flow Up, Work Flows Down

```
Human
  │
  │ (user story)
  ▼
Director ◄─────────────────────────────┐
  │                                    │
  │ (features)                         │ (questions,
  ▼                                    │  decisions)
Lead ◄────────────────────────┐        │
  │                           │        │
  │ (tasks)                   │        │
  ▼                           │        │
Worker ───────────────────────┴────────┘
  │
  │ (completion)
  ▼
(rolls back up)
```

When a worker has a question → Lead
When a lead has a question → Director
When director has a question → Human

This is the **escalation path**. Everything eventually gets answered.

---

## Dependencies Gate Concurrency

If Task B depends on Task A:
- Task B stays `blocked` until Task A is `done`
- Worker for Task B is not hired until unblocked
- This naturally limits active agents

```sql
-- A lead checks what can be started
SELECT * FROM work_items
WHERE parent_id = :my_feature
  AND status = 'pending'
  AND (depends_on IS NULL OR all_dependencies_complete(depends_on));
```

### System-Wide Limits

Config option:
```yaml
limits:
  max_active_workers: 5
  max_active_leads: 3
  max_total_agents: 10
```

When limit reached, new hires queue until a slot opens.

---

## Phase: Planning vs Execution

The same system handles both. The difference is **what tools are available**.

### Planning Phase

```go
// Available to all agents during planning
planningTools := []Tool{
    "create_work_item",    // Add to the plan
    "update_work_item",    // Refine description
    "add_dependency",      // Link items
    "raise_question",      // Need clarity
    "mark_ready",          // This item is fully defined
}
```

Output: A complete, hierarchical plan with all items in `ready` status.

### Execution Phase

```go
// Available during execution
executionTools := []Tool{
    "view_task",           // See what to do
    "bash",                // Run commands
    "edit",                // Modify files
    "read",                // Read files
    "mark_done",           // Complete the task
    "raise_blocker",       // Stuck, need help
}
```

Output: Completed work, code changes, artifacts.

---

## Meta Example: This System Plans Itself

Human input:
```
As a developer, I want a multi-agent orchestration system
so that I can delegate complex work to a hierarchy of AI agents.

Acceptance:
- [ ] Agents organized as Director → Lead → Worker
- [ ] Work broken down recursively
- [ ] Dependencies respected
- [ ] Questions escalate up
- [ ] Completion rolls up
```

Director receives this → Hires leads for:
- Lead 1: "Data Layer" (database schema, persistence)
- Lead 2: "Agent Lifecycle" (hire, fire, tmux management)
- Lead 3: "Communication" (message passing, escalation)
- Lead 4: "Work Decomposition" (breaking down items)

Each lead breaks their feature into tasks → Workers implement.

**The system builds itself.**

---

## Database Choice

Options considered:

| Option | Pros | Cons |
|--------|------|------|
| **SQLite** | Standard SQL, great tooling, human-readable | External dependency |
| **BoltDB** | Pure Go, embedded, fast | Less familiar, key-value only |
| **BadgerDB** | Pure Go, LSM-tree, fast | More complex |
| **JSON files** | Simple, debuggable | No queries, manual indexing |

Leaning toward **SQLite** because:
- Relational model fits the hierarchy
- Can query dependencies easily
- Human can inspect with any SQL tool
- Pure Go driver available (`modernc.org/sqlite`)

---

## Open Design Questions

1. **How does an agent know it's in planning vs execution phase?**
   - Explicit mode flag?
   - Different tool sets injected?
   - Work item metadata?

2. **When does planning end and execution begin?**
   - Human approval gate?
   - All items marked `ready`?
   - Director decision?

3. **Can planning and execution overlap?**
   - Start executing Feature A while still planning Feature B?
   - Or strict phases?

4. **How detailed should planning get?**
   - Down to file-level changes?
   - Or just "implement login form" and let worker figure it out?
