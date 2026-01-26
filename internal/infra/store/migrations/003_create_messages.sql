CREATE TABLE IF NOT EXISTS messages (
    id TEXT PRIMARY KEY,
    from_agent TEXT NOT NULL,
    to_agent TEXT NOT NULL,
    type TEXT NOT NULL,
    content TEXT NOT NULL,
    related_work TEXT,
    read BOOLEAN DEFAULT FALSE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    read_at DATETIME
);

CREATE INDEX IF NOT EXISTS idx_messages_to_unread ON messages(to_agent, read);
CREATE INDEX IF NOT EXISTS idx_messages_to_agent ON messages(to_agent, created_at);
