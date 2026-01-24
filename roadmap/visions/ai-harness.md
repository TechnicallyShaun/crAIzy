# AI Harness Vision

I need like a harness for the AIs.

Start the project again from a fresh, build it slower.

There needs to be a set process, but very generic so it can tackle dynamic workloads.

## Organisational Structure

I'm thinking like a company, with organisational units.
So we have a:
- Director
- Leads
- Workers

Top level manages,
Mid level breaks down tasks.
Bottom level does the work.

Could it follow a company process?

## Agent Lifecycle

| Role | Action | Description |
|------|--------|-------------|
| Director | Boss | Top-level orchestrator |
| | Hire | Pull from agent pool |
| | Idle | Park inactive agents |
| | Fire | Release agents |
| | | Brings in workers |

## Communication Layer

Needs a communication layer, the com layer needs to be outside the AIs,
Then a trigger event that whoever it's to, it gets posted in their chat.

---

## Build Order

1. Start with a UI, dashboard
2. Set a dynamic structure
3. Set the comm layer

## Flow

```
Human => "we need to do X" => Director
Director => task breakdown => Lead
Lead => breaks into smaller tasks => Worker
Worker => completes => Lead
Lead => needs a process to close out epics? Chunks?
Need a line back to human for questions TBD
```

## Self-Improvement Layer

Need a layer that continually improves itself, turn manual tasks into automations, to bring into capabilities.

Need to flesh out this flow, AI should help here.

## Tasks

- [ ] Discover flow of system
- [ ] Prototype tmux windows
- [ ] Decide comm layer/task management
