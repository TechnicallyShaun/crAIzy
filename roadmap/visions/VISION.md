# crAIzy Vision Document

> A generic, recursive AI orchestration harness that decomposes work hierarchically.

---

## What Is This?

A system that takes a body of work and autonomously:
1. Breaks it into smaller pieces
2. Delegates to specialized agents
3. Manages dependencies and concurrency
4. Rolls up completion
5. Escalates questions to humans

The same pattern applies whether you're **planning** (vision → features → tasks) or **executing** (tasks → code → done).

---

## Core Principles

### 1. Fractal Decomposition

Every level does the same thing: **big → smaller → delegate or execute**.

```
Vision → Features → Stories → Tasks → Execution
```

The depth isn't fixed. A simple task might be 2 levels. A complex project might be 6.

### 2. Hierarchy Models an Organization

| Role | Responsibility |
|------|----------------|
| **Director** | Receives high-level goals, hires leads per major chunk |
| **Lead** | Owns a feature/chunk, breaks it down, hires workers |
| **Worker** | Executes atomic tasks, signals completion |

Agents can query the org structure. Everyone knows who their parent is.

### 3. Tools as Guardrails

Agents don't act freely. They call **structured tools**:
- `hire_worker(task_id, agent_type)`
- `mark_done(summary)`
- `raise_question(content)`
- `create_work_item(parent_id, title, description)`

The Go harness executes these, updates the database, and logs everything.

### 4. Database as Source of Truth

All state lives in the database:
- Agent pool and active agents
- Work items (hierarchical)
- Messages between agents
- Audit log of all actions

Agents query the database to understand their context.

### 5. External Communication Layer

Agents don't talk directly. The harness:
1. Receives a tool call (e.g., `send_message`)
2. Records it in the database
3. Injects it into the recipient's tmux session

This decouples agents and creates a clean message trail.

---

## Concurrency Model

### Rule of Three (Parametrized)

A lead with 10 tasks doesn't spawn 10 workers. It spawns **N concurrent** (default: 3).

```
Lead has tasks: [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]

Active workers: [Worker-A (task 1), Worker-B (task 2), Worker-C (task 3)]

Worker-A completes → terminates
Active workers: [Worker-B, Worker-C, Worker-D (task 4)]

...and so on
```

Configurable via:
```yaml
limits:
  max_concurrent_workers_per_lead: 3
  max_total_agents: 10
```

### Dependencies Gate Work

If Task 5 depends on Task 3:
- Task 5 stays `blocked`
- Only becomes `pending` when Task 3 is `done`
- This naturally throttles concurrency

---

## Planning + Execution Overlap

These phases are **not strictly sequential**. They can interleave:

### Macro Level
- Define Feature A while building Feature B
- Plan long-term stages while executing early stages

### Micro Level (Branch-Based)
```
Feature X has 5 known tasks, 2 unclear tasks

Lead:
  - Hires workers for tasks 1-5 (known)
  - Raises questions for tasks 6-7 (unclear)
  - Work proceeds on a branch

Human answers questions:
  - Tasks 6-7 become defined
  - Workers hired for them
  - Branch continues

When all complete:
  - Lead rolls up
  - Branch ready for merge
```

This prevents blocking on unknowns while still getting clarity.

---

## Human Interface

Human provides **user stories** or **intent**:

```
As a user, I want to authenticate with Google
so that I don't need another password.

Acceptance:
- "Login with Google" button
- Redirect flow works
- Account created or linked
```

Human also:
- Answers escalated questions
- Monitors progress via dashboard
- Optionally approves at phase gates

---

## Startup Reconciliation

On app start, the harness:

1. Queries database for known sessions
2. Queries tmux for active sessions (by prefix)
3. Reconciles:
   - **Healthy**: In DB + in tmux → restore to UI
   - **Zombie**: In DB + not in tmux → mark dead in DB
   - **Orphan**: In tmux + not in DB → prompt user (adopt or kill)

This handles crashes gracefully.

---

## Self-Improvement Layer (Future)

Every action is logged. Over time:
1. Identify repeated patterns
2. Propose automations
3. Human approves → new capability

The system learns from its own execution.

---

## Tech Stack

| Component | Choice | Rationale |
|-----------|--------|-----------|
| **Language** | Go | Fast, single binary, good CLI tooling |
| **TUI** | Bubble Tea | Modern, composable, great UX |
| **Sessions** | tmux | Battle-tested, scriptable, survives crashes |
| **Database** | SQLite | Relational, queryable, human-inspectable |
| **Config** | YAML | Human-readable, easy to edit |

---

## What Exists Today

- Basic TUI scaffold (Bubble Tea)
- Agent config loading (AGENTS.yml)
- Modal system (agent selector, name input)
- Side menu placeholder
- Content area placeholder

**Not yet built:**
- SQLite persistence layer
- Agent lifecycle (hire/fire/idle)
- Work item management
- Message passing
- Tmux integration (was removed in reset)
- Reconciliation on startup

---

## Next Steps

1. **Document** - Capture this vision (done)
2. **Analyze** - Review current codebase structure
3. **Schema** - Define SQLite tables
4. **Foundation** - Build persistence layer
5. **Lifecycle** - Implement hire/fire flow
6. **Loop** - Get one Director → Lead → Worker cycle working
7. **Iterate** - Add features incrementally

---

## Meta Note

This system should be able to orchestrate its own development. The vision you're reading could be fed to a Director, decomposed into features, and built by the system itself.

We're doing that process manually right now. The goal is to automate it.
