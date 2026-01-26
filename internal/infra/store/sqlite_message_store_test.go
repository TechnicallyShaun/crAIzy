package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/TechnicallyShaun/crAIzy/internal/domain"
)

func createTestMessageStore(t *testing.T) (*SQLiteMessageStore, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "craizy-msg-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	dbPath := filepath.Join(tmpDir, "test.db")
	agentStore, err := NewSQLiteAgentStore(dbPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create agent store: %v", err)
	}

	messageStore := NewSQLiteMessageStore(agentStore.DB())

	cleanup := func() {
		agentStore.Close()
		os.RemoveAll(tmpDir)
	}

	return messageStore, cleanup
}

func TestSQLiteMessageStore_Save(t *testing.T) {
	store, cleanup := createTestMessageStore(t)
	defer cleanup()

	msg := &domain.Message{
		ID:        "msg-001",
		From:      "worker-001",
		To:        "lead-001",
		Type:      domain.MessageTypeQuestion,
		Content:   "Which auth library should I use?",
		Read:      false,
		CreatedAt: time.Now(),
	}

	err := store.Save(msg)
	if err != nil {
		t.Fatalf("failed to save message: %v", err)
	}

	// Verify it was stored
	retrieved, err := store.Get(msg.ID)
	if err != nil {
		t.Fatalf("failed to get message: %v", err)
	}
	if retrieved.ID != msg.ID {
		t.Errorf("expected ID %q, got %q", msg.ID, retrieved.ID)
	}
	if retrieved.From != msg.From {
		t.Errorf("expected From %q, got %q", msg.From, retrieved.From)
	}
	if retrieved.Type != domain.MessageTypeQuestion {
		t.Errorf("expected Type %q, got %q", domain.MessageTypeQuestion, retrieved.Type)
	}
}

func TestSQLiteMessageStore_SaveWithRelatedWork(t *testing.T) {
	store, cleanup := createTestMessageStore(t)
	defer cleanup()

	relatedWork := "feature-oauth"
	msg := &domain.Message{
		ID:          "msg-002",
		From:        "worker-001",
		To:          "lead-001",
		Type:        domain.MessageTypeCompletion,
		Content:     "OAuth feature complete",
		RelatedWork: &relatedWork,
		Read:        false,
		CreatedAt:   time.Now(),
	}

	err := store.Save(msg)
	if err != nil {
		t.Fatalf("failed to save message: %v", err)
	}

	retrieved, err := store.Get(msg.ID)
	if err != nil {
		t.Fatalf("failed to get message: %v", err)
	}
	if retrieved.RelatedWork == nil || *retrieved.RelatedWork != relatedWork {
		t.Errorf("expected RelatedWork %q, got %v", relatedWork, retrieved.RelatedWork)
	}
}

func TestSQLiteMessageStore_MarkRead(t *testing.T) {
	store, cleanup := createTestMessageStore(t)
	defer cleanup()

	msg := &domain.Message{
		ID:        "msg-003",
		From:      "worker-001",
		To:        "lead-001",
		Type:      domain.MessageTypeQuestion,
		Content:   "Test",
		Read:      false,
		CreatedAt: time.Now(),
	}
	_ = store.Save(msg)

	err := store.MarkRead(msg.ID)
	if err != nil {
		t.Fatalf("failed to mark read: %v", err)
	}

	retrieved, err := store.Get(msg.ID)
	if err != nil {
		t.Fatalf("failed to get message: %v", err)
	}
	if !retrieved.Read {
		t.Error("message should be marked as read")
	}
	if retrieved.ReadAt == nil {
		t.Error("ReadAt should be set when marking as read")
	}
}

func TestSQLiteMessageStore_ListUnread(t *testing.T) {
	store, cleanup := createTestMessageStore(t)
	defer cleanup()

	// Add messages - some read, some unread
	messages := []*domain.Message{
		{ID: "msg-1", From: "sender", To: "recipient-001", Type: domain.MessageTypeInfo, Content: "msg1", Read: false, CreatedAt: time.Now()},
		{ID: "msg-2", From: "sender", To: "recipient-001", Type: domain.MessageTypeInfo, Content: "msg2", Read: true, CreatedAt: time.Now()},
		{ID: "msg-3", From: "sender", To: "recipient-001", Type: domain.MessageTypeInfo, Content: "msg3", Read: false, CreatedAt: time.Now()},
		{ID: "msg-4", From: "sender", To: "recipient-002", Type: domain.MessageTypeInfo, Content: "msg4", Read: false, CreatedAt: time.Now()},
	}

	for _, msg := range messages {
		_ = store.Save(msg)
	}

	unread, err := store.ListUnread("recipient-001")
	if err != nil {
		t.Fatalf("failed to list unread: %v", err)
	}

	if len(unread) != 2 {
		t.Errorf("expected 2 unread messages, got %d", len(unread))
	}

	for _, msg := range unread {
		if msg.Read {
			t.Error("ListUnread returned a read message")
		}
		if msg.To != "recipient-001" {
			t.Errorf("wrong recipient: %s", msg.To)
		}
	}
}

func TestSQLiteMessageStore_List(t *testing.T) {
	store, cleanup := createTestMessageStore(t)
	defer cleanup()

	// Add messages with different timestamps
	now := time.Now()
	messages := []*domain.Message{
		{ID: "msg-1", From: "sender", To: "recipient-001", Type: domain.MessageTypeInfo, Content: "oldest", Read: false, CreatedAt: now.Add(-2 * time.Hour)},
		{ID: "msg-2", From: "sender", To: "recipient-001", Type: domain.MessageTypeInfo, Content: "middle", Read: true, CreatedAt: now.Add(-1 * time.Hour)},
		{ID: "msg-3", From: "sender", To: "recipient-001", Type: domain.MessageTypeInfo, Content: "newest", Read: false, CreatedAt: now},
		{ID: "msg-4", From: "sender", To: "recipient-002", Type: domain.MessageTypeInfo, Content: "other", Read: false, CreatedAt: now},
	}

	for _, msg := range messages {
		_ = store.Save(msg)
	}

	t.Run("list all messages for recipient", func(t *testing.T) {
		msgs, err := store.List("recipient-001", 0)
		if err != nil {
			t.Fatalf("failed to list: %v", err)
		}
		if len(msgs) != 3 {
			t.Errorf("expected 3 messages, got %d", len(msgs))
		}
	})

	t.Run("list with limit", func(t *testing.T) {
		msgs, err := store.List("recipient-001", 2)
		if err != nil {
			t.Fatalf("failed to list: %v", err)
		}
		if len(msgs) != 2 {
			t.Errorf("expected 2 messages, got %d", len(msgs))
		}
	})
}

func TestSQLiteMessageStore_Get(t *testing.T) {
	store, cleanup := createTestMessageStore(t)
	defer cleanup()

	msg := &domain.Message{
		ID:        "msg-get-test",
		From:      "worker-001",
		To:        "lead-001",
		Type:      domain.MessageTypeAssignment,
		Content:   "Implement OAuth",
		Read:      false,
		CreatedAt: time.Now(),
	}
	_ = store.Save(msg)

	retrieved, err := store.Get(msg.ID)
	if err != nil {
		t.Fatalf("failed to get message: %v", err)
	}
	if retrieved.Content != "Implement OAuth" {
		t.Errorf("expected Content %q, got %q", "Implement OAuth", retrieved.Content)
	}
	if retrieved.Type != domain.MessageTypeAssignment {
		t.Errorf("expected Type %q, got %q", domain.MessageTypeAssignment, retrieved.Type)
	}
}

func TestSQLiteMessageStore_GetNonExistent(t *testing.T) {
	store, cleanup := createTestMessageStore(t)
	defer cleanup()

	_, err := store.Get("non-existent-id")
	if err == nil {
		t.Error("expected error for non-existent message")
	}
}

func TestSQLiteMessageStore_UnreadCount(t *testing.T) {
	store, cleanup := createTestMessageStore(t)
	defer cleanup()

	messages := []*domain.Message{
		{ID: "msg-1", From: "sender", To: "recipient-001", Type: domain.MessageTypeInfo, Content: "msg1", Read: false, CreatedAt: time.Now()},
		{ID: "msg-2", From: "sender", To: "recipient-001", Type: domain.MessageTypeInfo, Content: "msg2", Read: true, CreatedAt: time.Now()},
		{ID: "msg-3", From: "sender", To: "recipient-001", Type: domain.MessageTypeInfo, Content: "msg3", Read: false, CreatedAt: time.Now()},
		{ID: "msg-4", From: "sender", To: "recipient-002", Type: domain.MessageTypeInfo, Content: "msg4", Read: false, CreatedAt: time.Now()},
	}

	for _, msg := range messages {
		_ = store.Save(msg)
	}

	count, err := store.UnreadCount("recipient-001")
	if err != nil {
		t.Fatalf("failed to count unread: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 unread, got %d", count)
	}
}

func TestSQLiteMessageStore_Persistence(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "craizy-msg-persist-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "persist.db")

	// Create store and add message
	agentStore1, err := NewSQLiteAgentStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create agent store: %v", err)
	}
	messageStore1 := NewSQLiteMessageStore(agentStore1.DB())

	msg := &domain.Message{
		ID:        "persistent-msg",
		From:      "worker",
		To:        "lead",
		Type:      domain.MessageTypeStatus,
		Content:   "Progress update",
		Read:      false,
		CreatedAt: time.Now(),
	}
	_ = messageStore1.Save(msg)
	agentStore1.Close()

	// Reopen and verify persistence
	agentStore2, err := NewSQLiteAgentStore(dbPath)
	if err != nil {
		t.Fatalf("failed to reopen agent store: %v", err)
	}
	defer agentStore2.Close()
	messageStore2 := NewSQLiteMessageStore(agentStore2.DB())

	retrieved, err := messageStore2.Get("persistent-msg")
	if err != nil {
		t.Fatalf("failed to get persisted message: %v", err)
	}
	if retrieved.Content != "Progress update" {
		t.Errorf("expected Content %q, got %q", "Progress update", retrieved.Content)
	}
}
