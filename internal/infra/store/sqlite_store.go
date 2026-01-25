package store

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/TechnicallyShaun/crAIzy/internal/domain"
	"github.com/TechnicallyShaun/crAIzy/internal/logging"
	_ "modernc.org/sqlite"
)

// SQLiteAgentStore implements IAgentStore with SQLite persistence.
type SQLiteAgentStore struct {
	db *sql.DB
}

// NewSQLiteAgentStore creates a new SQLite-backed agent store.
// It opens the database with WAL mode and runs migrations.
func NewSQLiteAgentStore(dbPath string) (*SQLiteAgentStore, error) {
	logging.Entry("dbPath", dbPath)
	db, err := sql.Open("sqlite", dbPath+"?_pragma=journal_mode(WAL)")
	if err != nil {
		logging.Error(err, "dbPath", dbPath)
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Run migrations
	if err := Migrate(db); err != nil {
		logging.Error(err, "action", "migrate")
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	logging.Info("SQLite store initialized, dbPath=%s", dbPath)
	return &SQLiteAgentStore{db: db}, nil
}

// Close closes the database connection.
func (s *SQLiteAgentStore) Close() error {
	logging.Entry()
	return s.db.Close()
}

// Add stores a new agent.
func (s *SQLiteAgentStore) Add(agent *domain.Agent) error {
	logging.Entry("agentID", agent.ID)
	_, err := s.db.Exec(`
		INSERT INTO agents (id, project, agent_type, name, command, work_dir, status, created_at, terminated_at, branch, base_branch)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, agent.ID, agent.Project, agent.AgentType, agent.Name, agent.Command, agent.WorkDir,
		string(agent.Status), agent.CreatedAt, agent.TerminatedAt, agent.Branch, agent.BaseBranch)
	if err != nil {
		logging.Error(err, "agentID", agent.ID)
		return fmt.Errorf("failed to insert agent: %w", err)
	}
	logging.Info("agent added to store, agentID=%s", agent.ID)
	return nil
}

// Remove deletes an agent by ID.
func (s *SQLiteAgentStore) Remove(id string) error {
	logging.Entry("id", id)
	_, err := s.db.Exec("DELETE FROM agents WHERE id = ?", id)
	if err != nil {
		logging.Error(err, "id", id)
		return fmt.Errorf("failed to delete agent: %w", err)
	}
	logging.Info("agent removed from store, id=%s", id)
	return nil
}

// List returns all stored agents.
func (s *SQLiteAgentStore) List() []*domain.Agent {
	logging.Entry()
	rows, err := s.db.Query(`
		SELECT id, project, agent_type, name, command, work_dir, status, created_at, terminated_at, branch, base_branch
		FROM agents
		ORDER BY created_at DESC
	`)
	if err != nil {
		logging.Error(err)
		return nil
	}
	defer rows.Close()

	var agents []*domain.Agent
	for rows.Next() {
		agent := &domain.Agent{}
		var status string
		var terminatedAt sql.NullTime
		var branch, baseBranch sql.NullString
		err := rows.Scan(
			&agent.ID, &agent.Project, &agent.AgentType, &agent.Name,
			&agent.Command, &agent.WorkDir, &status, &agent.CreatedAt, &terminatedAt,
			&branch, &baseBranch,
		)
		if err != nil {
			logging.Error(err, "action", "scan row")
			continue
		}
		agent.Status = domain.AgentStatus(status)
		if terminatedAt.Valid {
			agent.TerminatedAt = &terminatedAt.Time
		}
		if branch.Valid {
			agent.Branch = branch.String
		}
		if baseBranch.Valid {
			agent.BaseBranch = baseBranch.String
		}
		agents = append(agents, agent)
	}
	logging.Debug("listed %d agents from store", len(agents))
	return agents
}

// Get retrieves an agent by ID.
func (s *SQLiteAgentStore) Get(id string) *domain.Agent {
	logging.Entry("id", id)
	agent := &domain.Agent{}
	var status string
	var terminatedAt sql.NullTime
	var branch, baseBranch sql.NullString
	err := s.db.QueryRow(`
		SELECT id, project, agent_type, name, command, work_dir, status, created_at, terminated_at, branch, base_branch
		FROM agents WHERE id = ?
	`, id).Scan(
		&agent.ID, &agent.Project, &agent.AgentType, &agent.Name,
		&agent.Command, &agent.WorkDir, &status, &agent.CreatedAt, &terminatedAt,
		&branch, &baseBranch,
	)
	if err != nil {
		logging.Debug("agent not found, id=%s", id)
		return nil
	}
	agent.Status = domain.AgentStatus(status)
	if terminatedAt.Valid {
		agent.TerminatedAt = &terminatedAt.Time
	}
	if branch.Valid {
		agent.Branch = branch.String
	}
	if baseBranch.Valid {
		agent.BaseBranch = baseBranch.String
	}
	return agent
}

// Exists checks if an agent with the given ID exists.
func (s *SQLiteAgentStore) Exists(id string) bool {
	logging.Entry("id", id)
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM agents WHERE id = ?", id).Scan(&count)
	if err != nil {
		logging.Error(err, "id", id)
		return false
	}
	return count > 0
}

// UpdateStatus updates the status of an agent.
func (s *SQLiteAgentStore) UpdateStatus(id string, status domain.AgentStatus) error {
	logging.Entry("id", id, "status", status)
	var terminatedAt interface{}
	if status == domain.AgentStatusTerminated {
		now := time.Now()
		terminatedAt = now
	}

	_, err := s.db.Exec(`
		UPDATE agents SET status = ?, terminated_at = ? WHERE id = ?
	`, string(status), terminatedAt, id)
	if err != nil {
		logging.Error(err, "id", id, "status", status)
		return fmt.Errorf("failed to update agent status: %w", err)
	}
	logging.Info("agent status updated, id=%s, status=%s", id, status)
	return nil
}
