# Git Workflow Thoughts

Loose questions and ideas related to git integration that may inform future decisions.

## Merge Timing

Should `m` (merge) require the agent to be stopped first, or allow merging while agent is still running?

- **Allow mid-work merge**: Lets user checkpoint progress, agent can keep working and merge again later
- **Require stopped**: Cleaner mental model, no confusion about commits appearing after merge

## Visual Indicators

Should agents with unmerged commits show differently in the agent list?

- Could show a `*` or different color for agents with commits not yet merged
- Helps user track which agents have "finished" work ready for integration
- May require polling git status which could have performance implications

## Post-Merge Agent State

After successful merge, should the agent be automatically killed? Or left running for further iterations?

- **Auto-kill**: Clean workflow, merge = done
- **Keep running**: User might want to iterate further on same branch, or verify the merge before killing
- Could offer choice in the merge success modal: "Merged successfully. Kill agent? [Y/n]"

## Branch Cleanup Strategy

If users frequently create/kill agents with similar names, old branches could accumulate. Consider:

- Periodic cleanup prompt: "You have X old craizy branches. Clean up?"
- Automatic cleanup of merged branches on startup
- Namespace branches: `craizy/{name}` instead of just `{name}`
