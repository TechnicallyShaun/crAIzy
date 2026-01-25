-- Add git-related fields to agents table (SQLite doesn't support IF NOT EXISTS for columns)
-- We check if the column exists by trying to select it first
-- If this fails on fresh DBs, the columns are added by the CREATE TABLE in migration 001

-- For existing databases, we need to handle this programmatically in Go
