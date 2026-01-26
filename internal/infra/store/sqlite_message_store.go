package store

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/TechnicallyShaun/crAIzy/internal/domain"
	"github.com/TechnicallyShaun/crAIzy/internal/logging"
)

// SQLiteMessageStore implements IMessageStore with SQLite persistence.
type SQLiteMessageStore struct {
	db *sql.DB
}

// NewSQLiteMessageStore creates a new SQLite-backed message store.
// It uses an existing database connection (migrations are run by agent store init).
func NewSQLiteMessageStore(db *sql.DB) *SQLiteMessageStore {
	logging.Entry()
	return &SQLiteMessageStore{db: db}
}

// Save stores a new message.
func (s *SQLiteMessageStore) Save(msg *domain.Message) error {
	logging.Entry("msgID", msg.ID)
	_, err := s.db.Exec(`
		INSERT INTO messages (id, from_agent, to_agent, type, content, related_work, read, created_at, read_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, msg.ID, msg.From, msg.To, string(msg.Type), msg.Content, msg.RelatedWork,
		msg.Read, msg.CreatedAt, msg.ReadAt)
	if err != nil {
		logging.Error(err, "msgID", msg.ID)
		return fmt.Errorf("failed to insert message: %w", err)
	}
	logging.Info("message saved, msgID=%s", msg.ID)
	return nil
}

// MarkRead marks a message as read.
func (s *SQLiteMessageStore) MarkRead(id string) error {
	logging.Entry("id", id)
	now := time.Now()
	_, err := s.db.Exec(`
		UPDATE messages SET read = TRUE, read_at = ? WHERE id = ?
	`, now, id)
	if err != nil {
		logging.Error(err, "id", id)
		return fmt.Errorf("failed to mark message as read: %w", err)
	}
	logging.Info("message marked as read, id=%s", id)
	return nil
}

// ListUnread returns all unread messages for a recipient.
func (s *SQLiteMessageStore) ListUnread(recipientID string) ([]*domain.Message, error) {
	logging.Entry("recipientID", recipientID)
	rows, err := s.db.Query(`
		SELECT id, from_agent, to_agent, type, content, related_work, read, created_at, read_at
		FROM messages
		WHERE to_agent = ? AND read = FALSE
		ORDER BY created_at ASC
	`, recipientID)
	if err != nil {
		logging.Error(err, "recipientID", recipientID)
		return nil, fmt.Errorf("failed to list unread messages: %w", err)
	}
	defer rows.Close()

	return s.scanMessages(rows)
}

// List returns messages for a recipient with a limit (0 = no limit).
func (s *SQLiteMessageStore) List(recipientID string, limit int) ([]*domain.Message, error) {
	logging.Entry("recipientID", recipientID, "limit", limit)

	var query string
	var args []interface{}

	if limit > 0 {
		query = `
			SELECT id, from_agent, to_agent, type, content, related_work, read, created_at, read_at
			FROM messages
			WHERE to_agent = ?
			ORDER BY created_at DESC
			LIMIT ?
		`
		args = []interface{}{recipientID, limit}
	} else {
		query = `
			SELECT id, from_agent, to_agent, type, content, related_work, read, created_at, read_at
			FROM messages
			WHERE to_agent = ?
			ORDER BY created_at DESC
		`
		args = []interface{}{recipientID}
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		logging.Error(err, "recipientID", recipientID)
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}
	defer rows.Close()

	return s.scanMessages(rows)
}

// Get retrieves a message by ID.
func (s *SQLiteMessageStore) Get(id string) (*domain.Message, error) {
	logging.Entry("id", id)
	msg := &domain.Message{}
	var msgType string
	var relatedWork sql.NullString
	var readAt sql.NullTime

	err := s.db.QueryRow(`
		SELECT id, from_agent, to_agent, type, content, related_work, read, created_at, read_at
		FROM messages WHERE id = ?
	`, id).Scan(
		&msg.ID, &msg.From, &msg.To, &msgType, &msg.Content,
		&relatedWork, &msg.Read, &msg.CreatedAt, &readAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			logging.Debug("message not found, id=%s", id)
			return nil, fmt.Errorf("message not found: %s", id)
		}
		logging.Error(err, "id", id)
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	msg.Type = domain.MessageType(msgType)
	if relatedWork.Valid {
		msg.RelatedWork = &relatedWork.String
	}
	if readAt.Valid {
		msg.ReadAt = &readAt.Time
	}

	return msg, nil
}

// UnreadCount returns the count of unread messages for a recipient.
func (s *SQLiteMessageStore) UnreadCount(recipientID string) (int, error) {
	logging.Entry("recipientID", recipientID)
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM messages WHERE to_agent = ? AND read = FALSE
	`, recipientID).Scan(&count)
	if err != nil {
		logging.Error(err, "recipientID", recipientID)
		return 0, fmt.Errorf("failed to count unread messages: %w", err)
	}
	return count, nil
}

// scanMessages scans rows into a slice of Message pointers.
func (s *SQLiteMessageStore) scanMessages(rows *sql.Rows) ([]*domain.Message, error) {
	var messages []*domain.Message
	for rows.Next() {
		msg := &domain.Message{}
		var msgType string
		var relatedWork sql.NullString
		var readAt sql.NullTime

		err := rows.Scan(
			&msg.ID, &msg.From, &msg.To, &msgType, &msg.Content,
			&relatedWork, &msg.Read, &msg.CreatedAt, &readAt,
		)
		if err != nil {
			logging.Error(err, "action", "scan message row")
			continue
		}

		msg.Type = domain.MessageType(msgType)
		if relatedWork.Valid {
			msg.RelatedWork = &relatedWork.String
		}
		if readAt.Valid {
			msg.ReadAt = &readAt.Time
		}

		messages = append(messages, msg)
	}
	return messages, nil
}
