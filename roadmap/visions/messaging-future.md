# Messaging - Future Considerations

Future enhancements to the messaging system, deferred from MVP.

## Threading

Add a `thread` field to messages enabling conversation threads.

```go
type Message struct {
    // ... existing fields ...
    Thread *string  // messageId of thread root (null = new thread)
}
```

**Design questions to resolve:**
- Thread root: `thread = null` or self-reference?
- Thread metadata (subject line for entire thread)?
- Query patterns: list threads vs list messages in thread
- UI: threaded view vs flat chronological

**Schema addition:**
```sql
ALTER TABLE messages ADD COLUMN thread_id TEXT REFERENCES messages(id);
CREATE INDEX idx_messages_thread ON messages(thread_id);
```

## Reserved "Director" Recipient

Make "director" a reserved recipient ID like "human".

- Director is the orchestrating agent that manages leads/workers
- Messages to "director" are routed to the active director agent
- Allows agents to escalate without knowing the specific director instance

**Implementation:**
- Add to reserved IDs: `human`, `director`
- Resolve `director` to actual tmux session at delivery time
- Handle case where no director is active (queue like human)

## Thread Property Inheritance

When replying to a message, auto-inherit properties from parent:
- `relatedWork` carries forward unless explicitly overridden
- Possibly `type` (answer inherits context of question)

Deferred as it adds complexity and the simple model works for MVP.
