package store

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/TechnicallyShaun/crAIzy/internal/domain"
	_ "modernc.org/sqlite"
)

// SQLiteAgentStore implements IAgentStore with SQLite persistence.
type SQLiteAgentStore struct {
	db *sql.DB
}

// NewSQLiteAgentStore creates a new SQLite-backed agent store.
// It opens the database with WAL mode and runs migrations.
func NewSQLiteAgentStore(dbPath string) (*SQLiteAgentStore, error) {
	db, err := sql.Open("sqlite", dbPath+"?_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Run migrations
	if err := Migrate(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return &SQLiteAgentStore{db: db}, nil
}

// Close closes the database connection.
func (s *SQLiteAgentStore) Close() error {
	return s.db.Close()
}

// Add stores a new agent.
func (s *SQLiteAgentStore) Add(agent *domain.Agent) error {
	_, err := s.db.Exec(`
		INSERT INTO agents (id, project, agent_type, name, command, work_dir, status, created_at, terminated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, agent.ID, agent.Project, agent.AgentType, agent.Name, agent.Command, agent.WorkDir,
		string(agent.Status), agent.CreatedAt, agent.TerminatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert agent: %w", err)
	}
	return nil
}

// Remove deletes an agent by ID.
func (s *SQLiteAgentStore) Remove(id string) error {
	_, err := s.db.Exec("DELETE FROM agents WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete agent: %w", err)
	}
	return nil
}

// List returns all stored agents.
func (s *SQLiteAgentStore) List() []*domain.Agent {
	rows, err := s.db.Query(`
		SELECT id, project, agent_type, name, command, work_dir, status, created_at, terminated_at
		FROM agents
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var agents []*domain.Agent
	for rows.Next() {
		agent := &domain.Agent{}
		var status string
		var terminatedAt sql.NullTime
		err := rows.Scan(
			&agent.ID, &agent.Project, &agent.AgentType, &agent.Name,
			&agent.Command, &agent.WorkDir, &status, &agent.CreatedAt, &terminatedAt,
		)
		if err != nil {
			continue
		}
		agent.Status = domain.AgentStatus(status)
		if terminatedAt.Valid {
			agent.TerminatedAt = &terminatedAt.Time
		}
		agents = append(agents, agent)
	}
	return agents
}

// Get retrieves an agent by ID.
func (s *SQLiteAgentStore) Get(id string) *domain.Agent {
	agent := &domain.Agent{}
	var status string
	var terminatedAt sql.NullTime
	err := s.db.QueryRow(`
		SELECT id, project, agent_type, name, command, work_dir, status, created_at, terminated_at
		FROM agents WHERE id = ?
	`, id).Scan(
		&agent.ID, &agent.Project, &agent.AgentType, &agent.Name,
		&agent.Command, &agent.WorkDir, &status, &agent.CreatedAt, &terminatedAt,
	)
	if err != nil {
		return nil
	}
	agent.Status = domain.AgentStatus(status)
	if terminatedAt.Valid {
		agent.TerminatedAt = &terminatedAt.Time
	}
	return agent
}

// Exists checks if an agent with the given ID exists.
func (s *SQLiteAgentStore) Exists(id string) bool {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM agents WHERE id = ?", id).Scan(&count)
	if err != nil {
		return false
	}
	return count > 0
}

// UpdateStatus updates the status of an agent.
func (s *SQLiteAgentStore) UpdateStatus(id string, status domain.AgentStatus) error {
	var terminatedAt interface{}
	if status == domain.AgentStatusTerminated {
		now := time.Now()
		terminatedAt = now
	}

	_, err := s.db.Exec(`
		UPDATE agents SET status = ?, terminated_at = ? WHERE id = ?
	`, string(status), terminatedAt, id)
	if err != nil {
		return fmt.Errorf("failed to update agent status: %w", err)
	}
	return nil
}
