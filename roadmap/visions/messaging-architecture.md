# Messaging Architecture: Intent-Driven Routing

## Core Principle

**Agents express intent, the system handles routing.**

An agent shouldn't say "send this to lead-001". They should say "I have a question" or "I'm blocked" or "task complete". The messaging system knows the hierarchy, knows the work structure, and routes accordingly.

This inverts the current design. Instead of:
```
agent.sendTo("lead-001", "question", "Which auth library?")
```

It becomes:
```
agent.signal(Question, "Which auth library?")
// System knows: worker's parent is lead-001, route there
// If lead can't answer, escalates to director
// If director can't answer, escalates to human
```

The agent is **decoupled from topology**. Hierarchy can change, routing rules can evolve, agents don't care.

---

## Signal Types (Intent Taxonomy)

| Signal | Meaning | Blocking? | System Action |
|--------|---------|-----------|---------------|
| `Question` | I'd like to know something | No | Escalate, agent continues working |
| `Blocked` | I cannot continue | **Yes** | Escalate, mark work blocked, agent waits |
| `NeedDecision` | Multiple valid paths, need human choice | **Yes** | Escalate to human, agent waits |
| `Status` | Progress update, FYI | No | Broadcast to interested parties |
| `Complete` | My work is done | No | Notify parent, trigger rollup |
| `Failed` | I tried but couldn't succeed | No | Escalate, mark work failed |

**Blocking vs Non-blocking**:
- `Question` - "Which auth library do you prefer?" - Agent can proceed with a reasonable default, answer informs future work
- `Blocked` - "I don't have API credentials" - Agent literally cannot continue
- `NeedDecision` - "Should we use OAuth or JWT?" - Architectural choice, wrong pick wastes significant work

Only `Blocked` and `NeedDecision` halt the agent. Everything else flows async.

---

## The Problem

In a Director -> Lead -> Worker hierarchy:
- Who should workers talk to?
- How do questions escalate?
- How does status roll up?
- How does the human stay informed without being overwhelmed?

The current design assumes agents know exactly who to message. But:
- A worker shouldn't need to know if a question is for their lead or for human
- Leads shouldn't have to manually route every message
- The human shouldn't see every low-level status update

## Approach 1: Blackboard Pattern

**Concept**: Shared state space that interested parties check/poll.

```
┌─────────────────────────────────────────┐
│              BLACKBOARD                 │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐   │
│  │ Topic A │ │ Topic B │ │ Topic C │   │
│  └─────────┘ └─────────┘ └─────────┘   │
└─────────────────────────────────────────┘
      ▲    ▲        ▲           ▲
      │    │        │           │
   Worker Lead   Director    Human
   (write) (read/write) (read/write) (read)
```

**How it works**:
- Agents write to topics on the blackboard (e.g., "feature-auth/questions", "feature-auth/status")
- Interested parties poll or subscribe to topics they care about
- No direct addressing - you post, others discover

**Pros**:
- Decoupled - writers don't need to know readers
- Natural aggregation - leads see all their workers' posts
- Scales well - add new observers without changing producers
- Good for status/progress (many writers, few readers)

**Cons**:
- Polling overhead or complex subscription management
- Harder to ensure delivery (who's responsible for checking?)
- Questions may go unanswered (no clear owner)
- Less clear escalation path

**Best for**: Status updates, progress tracking, shared state

**Implementation sketch**:
```go
type BlackboardEntry struct {
    Topic     string    // "feature-auth/questions"
    Author    string
    Content   string
    CreatedAt time.Time
}

type IBlackboard interface {
    Post(topic, author, content string) error
    Read(topicPattern string, since time.Time) ([]*BlackboardEntry, error)
    Watch(topicPattern string) <-chan *BlackboardEntry
}
```

---

## Approach 2: Work/Task-Scoped PubSub

**Concept**: Messages are published to work items, and hierarchy subscribes based on ownership.

```
                    ┌──────────────┐
                    │  Epic: MVP   │ ◄── Director subscribes
                    └──────────────┘
                           │
              ┌────────────┼────────────┐
              ▼            ▼            ▼
       ┌──────────┐ ┌──────────┐ ┌──────────┐
       │Feature A │ │Feature B │ │Feature C │ ◄── Leads subscribe
       └──────────┘ └──────────┘ └──────────┘
              │
      ┌───────┼───────┐
      ▼       ▼       ▼
   ┌─────┐ ┌─────┐ ┌─────┐
   │Task1│ │Task2│ │Task3│ ◄── Workers subscribe
   └─────┘ └─────┘ └─────┘
```

**How it works**:
- Every message is scoped to a work item (task, feature, epic)
- Agents subscribe to work items they own
- Messages "bubble up" automatically through work hierarchy
- A worker posts to their task, lead sees it (owns parent feature), director sees it (owns parent epic)

**Pros**:
- Natural alignment with work hierarchy
- Automatic context - messages tied to work
- Built-in escalation through work tree
- Easy filtering - only see what's relevant to your work

**Cons**:
- Requires work item system to be robust
- Cross-cutting messages (not work-specific) need special handling
- What about human? Human owns everything? Or special subscription?

**Best for**: Status updates, completion signals, work-scoped questions

**Implementation sketch**:
```go
type WorkMessage struct {
    ID        string
    WorkItem  string    // "task-123", "feature-auth"
    Author    string
    Type      string    // question, status, completion
    Content   string
    CreatedAt time.Time
}

type IWorkPubSub interface {
    Publish(workItem, author, msgType, content string) error
    Subscribe(workItemPattern string, handler func(*WorkMessage))
    // Pattern: "epic-mvp/**" gets all children
}

// Automatic subscription based on work ownership
// Lead owns "feature-auth" -> auto-subscribed to "feature-auth/**"
// Director owns "epic-mvp" -> auto-subscribed to "epic-mvp/**"
```

---

## Approach 3: Director-Routed (Centralized)

**Concept**: All messages go to Director (or your parent), who decides routing.

```
                         ┌──────────┐
                    ┌───►│ Director │◄───┐
                    │    └──────────┘    │
                    │         │          │
            routes to    routes to   routes to
            human        lead        lead
                    │         │          │
                    ▼         ▼          ▼
              ┌───────┐  ┌────────┐  ┌────────┐
              │ Human │  │ Lead A │  │ Lead B │
              └───────┘  └────────┘  └────────┘
                              │
                    ┌─────────┼─────────┐
                    ▼         ▼         ▼
              ┌────────┐ ┌────────┐ ┌────────┐
              │Worker 1│ │Worker 2│ │Worker 3│
              └────────┘ └────────┘ └────────┘
                    │
                    └──► Lead A (always route to parent)
```

**How it works**:
- Workers always message their Lead
- Leads always message Director
- Director decides: handle it, route to another lead, or escalate to human
- Messages have intent ("question", "blocked", "decision-needed") that informs routing

**Pros**:
- Simple mental model for agents - just talk to parent
- Intelligent routing at decision points
- Director can batch/summarize for human
- Director has full visibility

**Cons**:
- Director becomes bottleneck
- Latency - every message goes through intermediary
- Director needs sophisticated routing logic
- Single point of failure

**Best for**: Questions, escalations, decisions

**Implementation sketch**:
```go
type RoutedMessage struct {
    ID        string
    From      string
    Intent    string    // question, blocked, decision, info
    Content   string
    CreatedAt time.Time
}

// Agent just sends to parent
func (w *Worker) Ask(question string) {
    msg := &RoutedMessage{From: w.ID, Intent: "question", Content: question}
    w.parent.Receive(msg)
}

// Parent decides routing
func (l *Lead) Receive(msg *RoutedMessage) {
    switch {
    case l.canAnswer(msg):
        l.reply(msg)
    case msg.Intent == "decision":
        l.escalateToDirector(msg)
    default:
        l.handleLocally(msg)
    }
}

func (d *Director) Receive(msg *RoutedMessage) {
    switch {
    case d.canAnswer(msg):
        d.reply(msg)
    case msg.Intent == "decision" || d.needsHuman(msg):
        d.escalateToHuman(msg)
    default:
        d.routeToLead(msg)
    }
}
```

---

## Approach 4: Hybrid (Recommended)

**Concept**: Use different patterns for different message types.

| Message Type | Pattern | Rationale |
|--------------|---------|-----------|
| **Questions/Escalations** | Director-routed | Need intelligent routing, human involvement |
| **Status/Progress** | Work-scoped PubSub | Many writers, few readers, tied to work |
| **Completion signals** | Work-scoped PubSub | Natural rollup through work hierarchy |
| **Shared state** | Blackboard | Read by many, updated incrementally |
| **Direct commands** | Point-to-point | Parent -> child assignments |

```
┌─────────────────────────────────────────────────────────────┐
│                    MESSAGE ROUTER                           │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │  Questions  │  │   Status    │  │   Assignments       │ │
│  │  Decisions  │  │  Progress   │  │   Completions       │ │
│  │             │  │             │  │                     │ │
│  │ → Director  │  │ → Work      │  │ → Direct to         │ │
│  │   Routed    │  │   PubSub    │  │   recipient         │ │
│  └─────────────┘  └─────────────┘  └─────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

**How it works**:
- Message type determines routing strategy
- Questions/decisions always escalate through hierarchy
- Status broadcasts to work-item subscribers
- Assignments go direct (parent knows child)

**Implementation sketch**:
```go
func (s *MessageService) Send(msg *Message) error {
    switch msg.Type {
    case "question", "decision", "blocked":
        return s.routeThroughHierarchy(msg)
    case "status", "progress":
        return s.publishToWorkItem(msg)
    case "assignment", "completion":
        return s.sendDirect(msg)
    default:
        return s.sendDirect(msg)
    }
}
```

---

## Decision Matrix

| Criteria | Blackboard | Work PubSub | Director-Routed | Hybrid |
|----------|------------|-------------|-----------------|--------|
| Simplicity | Medium | Medium | High | Low |
| Scalability | High | High | Low (bottleneck) | High |
| Guaranteed delivery | Low | Medium | High | High |
| Human isolation | Low | Medium | High | High |
| Latency | Low | Low | High | Varies |
| Implementation effort | Medium | High | Low | High |

---

## Open Questions

1. **Should "director" be a reserved ID like "human"?** This would let workers/leads send to "director" without knowing the instance.

2. **How does the hybrid router know work hierarchy?** Needs access to work item parent-child relationships.

3. **What happens when Director is down?** Queue messages? Fail? Escalate directly to human?

4. **Should status updates be push or pull?** Director polling vs. leads pushing.

5. **Message batching for human?** Director could summarize: "3 questions pending, 2 blockers, 5 status updates"

---

## Recommendation: Intent-Driven Routing

The approaches above are **implementation details**. The API agents see should be intent-based:

```go
// What agents see (simple, intent-based)
type Signal interface {
    Ask(question string)           // I need an answer
    Blocked(reason string)         // I can't proceed
    Status(update string)          // FYI, here's where I am
    Complete(summary string)       // I'm done with my work
    NeedDecision(options string)   // Human needs to choose
}

// What the system does (hidden from agents)
func (s *Router) Handle(signal Signal, from Agent) {
    switch signal.Intent() {
    case Question, Blocked, NeedDecision:
        // Escalate through hierarchy until answered
        s.escalate(signal, from.Parent())
    case Status:
        // Broadcast to work-item subscribers
        s.broadcast(signal, from.CurrentWork())
    case Complete:
        // Notify parent, update work state
        s.notifyParent(signal, from)
        s.markWorkDone(from.CurrentWork())
    }
}
```

**For MVP**: Implement the escalation path first (Question, Blocked, NeedDecision). These are the critical ones - they're what causes work to stall.

Status and Complete can start as simple parent notifications and evolve into pubsub later.

The key: **agents never specify recipients**. They signal intent, system routes.

---

## Related Documents

- `features/messaging.md` - Current point-to-point design
- `visions/messaging-future.md` - Threading and other enhancements
- `visions/architecture-deep-dive.md` - Hierarchy design
