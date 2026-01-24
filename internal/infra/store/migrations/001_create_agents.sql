CREATE TABLE IF NOT EXISTS agents (
    id TEXT PRIMARY KEY,
    project TEXT NOT NULL,
    agent_type TEXT NOT NULL,
    name TEXT NOT NULL,
    command TEXT NOT NULL,
    work_dir TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    created_at DATETIME NOT NULL,
    terminated_at DATETIME
);

CREATE INDEX IF NOT EXISTS idx_agents_project ON agents(project);
CREATE INDEX IF NOT EXISTS idx_agents_status ON agents(status);
