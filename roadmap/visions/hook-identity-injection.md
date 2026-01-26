# Hook-Based Identity Injection

## Problem

When crAIzy spawns an agent (Claude, Gemini, Copilot), that agent is a blank slate. It doesn't know:
- What role it plays (Director, Lead, Worker)
- Who it reports to
- What work it's assigned
- How to communicate with the system

We need to inject this identity **automatically** when the agent boots, before the human (or system) sends any commands.

---

## The Hook Opportunity

All three target CLIs support hooks that fire on session start:

| CLI | Event | Config Location |
|-----|-------|-----------------|
| Claude Code | `SessionStart` | `.claude/settings.json` |
| Gemini CLI | `SessionStart` | `.gemini/settings.json` |
| Copilot CLI | `sessionStart` | `.github/hooks/hooks.json` |

These hooks can:
1. Run a shell command
2. Output text that gets injected into the agent's context
3. Access environment variables from the parent process

---

## Options for Passing Identity

### Option A: Environment Variables (Multiple)

Orchestrator sets multiple env vars before spawning:

```go
cmd.Env = append(os.Environ(),
    "AGENT_ROLE=worker",
    "AGENT_REPORTS_TO=lead.a",
    "AGENT_ID=worker.2",
    "WORK_ITEM=story.5",
)
```

Hook script reads each:
```bash
ROLE="${AGENT_ROLE:-worker}"
cat ".roles/${ROLE}.md"
```

**Pros:**
- Explicit, no parsing needed
- Easy to add new variables

**Cons:**
- Multiple values to manage
- Can get out of sync
- Verbose spawning code

---

### Option B: Structured ID (Single Value)

Orchestrator sets one env var containing a structured identifier:

```go
// Format: hierarchy/path:work_item
cmd.Env = append(os.Environ(),
    "AGENT_ID=director/lead.a/worker.2:story.5",
)
```

Hook script parses it:
```bash
ID="${AGENT_ID}"
HIERARCHY="${ID%:*}"           # director/lead.a/worker.2
WORK_ITEM="${ID#*:}"           # story.5
ROLE=$(basename "$HIERARCHY" | sed 's/[.0-9]*$//')  # worker
PARENT=$(dirname "$HIERARCHY") # director/lead.a
```

**Pros:**
- Single source of truth
- ID is self-documenting
- ID = tmux session name (no translation)
- Easy to grep logs
- Hierarchy visible in the identifier
- Database queries: `WHERE agent_id LIKE 'lead.a/%'`

**Cons:**
- Requires parsing logic
- Format is a contract to maintain

---

### Option C: Session Name Convention

Use the tmux session name itself as the identifier (no env var needed):

```go
sessionID := "director/lead.a/worker.2:story.5"
cmd := exec.Command("tmux", "new-session", "-d", "-s", sessionID, ...)
```

Hook queries tmux for its own session name:
```bash
ID=$(tmux display-message -p '#{session_name}')
# ... parse as above
```

**Pros:**
- No env vars needed
- Session name is always available
- Natural integration with tmux

**Cons:**
- Couples hook to tmux (breaks if running without tmux)
- Session name has character restrictions

---

### Option D: Identity File (Written Before Spawn)

Orchestrator writes identity to a temp file, hook reads it:

```go
identityPath := fmt.Sprintf("/tmp/craizy-%s.identity", agentID)
os.WriteFile(identityPath, []byte(identityJSON), 0600)
cmd.Env = append(os.Environ(), "AGENT_IDENTITY_FILE="+identityPath)
```

Hook reads:
```bash
IDENTITY=$(cat "${AGENT_IDENTITY_FILE}")
# Parse JSON with jq
```

**Pros:**
- Can include complex/structured data
- No parsing in bash (use jq)

**Cons:**
- Temp file management
- Cleanup complexity
- Race conditions possible

---

### Option E: Query Database on Boot

Hook queries the crAIzy database directly:

```bash
ID="${AGENT_ID}"
IDENTITY=$(craizy agent info "$ID" --json)
# Parse JSON
```

**Pros:**
- Database is single source of truth
- Rich context available
- No stale data

**Cons:**
- Requires craizy CLI to be available
- Adds startup latency
- Hook depends on external tool

---

## Recommended Approach: Option B (Structured ID)

For crAIzy's hierarchical model, the **structured ID** approach is optimal:

```
director/lead.a/worker.2:story.5
│        │       │        │
│        │       │        └── work assignment
│        │       └── this agent (worker instance 2)
│        └── parent (lead instance a)
└── grandparent (director)
```

### Why This Fits crAIzy

1. **Hierarchy is explicit** - The ID *is* the org chart path
2. **Work assignment included** - Agent knows what to do immediately
3. **tmux naming** - Session name = ID (easy to find/kill)
4. **Logging** - One string captures full context
5. **Database** - Can query subtrees with LIKE patterns
6. **Escalation path** - Parent is derivable from ID

### Implementation

**Env var:** `AGENT_ID=director/lead.a/worker.2:story.5`

**Hook output:**
```markdown
# Role: Worker

You are a Worker agent in the crAIzy orchestration system.

## Your Identity
- Agent ID: director/lead.a/worker.2
- Role: Worker
- Reports To: director/lead.a (Lead)
- Assignment: story.5

## Your Responsibilities
- Execute the assigned work item directly
- Ask questions when blocked (escalate to Lead)
- Report completion when done

## Your Constraints
- Do NOT decompose work (that's Lead's job)
- Do NOT make architectural decisions (escalate)
- Do NOT spawn other agents

## Getting Started
Retrieve your assignment details:
craizy work get story.5
```

---

## Message Injection Methods

Once we have identity, how does it reach the agent?

### Method 1: Hook stdout (Recommended for Claude/Gemini)

Hook outputs text → CLI injects into context.

```json
// .claude/settings.json
{
  "hooks": {
    "SessionStart": [{
      "type": "command",
      "command": ".craizy/hooks/inject-identity.sh"
    }]
  }
}
```

**Pros:** Clean, no user-visible message
**Cons:** Relies on CLI's injection behavior

### Method 2: tmux send-keys (Fallback)

Orchestrator sends initial message after boot:

```go
// Wait for agent to be ready
time.Sleep(2 * time.Second)
tmux.SendKeys(sessionID, "Read your identity from AGENT_ID env var and begin.", "Enter")
```

**Pros:** Works regardless of CLI hook support
**Cons:** Visible in session, timing-dependent

### Method 3: CLAUDE.md / GEMINI.md / copilot-instructions.md

Static context file that references the dynamic identity:

```markdown
# Agent Instructions

You are a crAIzy agent. Your identity is in the AGENT_ID environment variable.

Parse it as: `hierarchy:work_item`

On startup:
1. Read $AGENT_ID
2. Determine your role from the hierarchy
3. Fetch your work assignment
4. Begin execution
```

**Pros:** No hook script needed
**Cons:** Agent must "understand" to parse; less reliable

---

## Hook Installation

Hooks must be installed per-CLI. This should happen during `craizy init`:

```
craizy init
├── Creates .craizy/
├── Creates .craizy/hooks/inject-identity.sh
├── Creates .craizy/roles/director.md
├── Creates .craizy/roles/lead.md
├── Creates .craizy/roles/worker.md
├── Creates .claude/settings.json (with hooks)
├── Creates .gemini/settings.json (with hooks)
└── Creates .github/hooks/hooks.json (with hooks)
```

### Idempotency

Init must be safe to re-run:
- Check if hook config exists
- Merge crAIzy hooks with existing user hooks
- Don't overwrite user customizations

---

## Open Questions

1. **Separator choice** - `/` vs `-` for hierarchy? `/` is path-like but may conflict with filesystem. `-` is safe but less readable.

2. **Work item in ID** - Should work item be part of the ID or separate env var? Including it makes the ID fully self-describing but longer.

3. **Agent instance numbering** - `worker.2` vs `worker-alpha` vs UUID? Numbers are simple but UUIDs avoid collision.

4. **Hook failure handling** - What if the hook fails? Agent boots without identity? Retry? Block?

5. **Multi-CLI support** - Can one hook script work for all CLIs, or do we need CLI-specific scripts?

---

## Related Documents

- `visions/hierarchical-agents.md` - Org structure
- `visions/process-model.md` - Work decomposition
- `features/complete/craizy-init.md` - Current init implementation
