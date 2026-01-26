# Git-Friendly Database Migration

## Current State

Using SQLite (`modernc.org/sqlite`) with database file at `~/.craizy/craizy.db`.

SQLite is fine for runtime state but isn't git-friendly:
- Binary file that doesn't diff
- Can't merge changes
- No built-in versioning/history

## Why Git-Friendly?

A git-native database would enable:
- Version control of data alongside code
- Branch-per-feature data isolation
- Time-travel queries (audit trail for free)
- Merge data changes between branches
- Human-readable diffs of data

## Options to Explore

### Dolt

MySQL-compatible database with git semantics built-in.
- `dolt commit`, `dolt branch`, `dolt merge`
- SQL interface (same Go `database/sql` patterns)
- Go native via `github.com/dolthub/dolt/go`

### Beads

(Research needed - noted as potential option)

## Decision

Not yet decided. Need to evaluate:
- Performance overhead
- Complexity of integration
- Whether git semantics are worth it for our use case

## Architecture Note

Current `IAgentStore` and `IMessageStore` interfaces should make swapping implementations straightforward - this was intentionally designed as noted in persistence-layer.md.
