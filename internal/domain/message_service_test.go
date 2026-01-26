package domain

import (
	"testing"
)

// Mock message store
type mockMessageStore struct {
	messages    map[string]*Message
	saveErr     error
	markReadErr error
	getErr      error
}

func newMockMessageStore() *mockMessageStore {
	return &mockMessageStore{messages: make(map[string]*Message)}
}

func (m *mockMessageStore) Save(msg *Message) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.messages[msg.ID] = msg
	return nil
}

func (m *mockMessageStore) MarkRead(id string) error {
	if m.markReadErr != nil {
		return m.markReadErr
	}
	if msg, ok := m.messages[id]; ok {
		msg.Read = true
	}
	return nil
}

func (m *mockMessageStore) ListUnread(recipientID string) ([]*Message, error) {
	var msgs []*Message
	for _, msg := range m.messages {
		if msg.To == recipientID && !msg.Read {
			msgs = append(msgs, msg)
		}
	}
	return msgs, nil
}

func (m *mockMessageStore) List(recipientID string, limit int) ([]*Message, error) {
	var msgs []*Message
	for _, msg := range m.messages {
		if msg.To == recipientID {
			msgs = append(msgs, msg)
			if limit > 0 && len(msgs) >= limit {
				break
			}
		}
	}
	return msgs, nil
}

func (m *mockMessageStore) Get(id string) (*Message, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	msg, ok := m.messages[id]
	if !ok {
		return nil, &messageNotFoundError{id: id}
	}
	return msg, nil
}

func (m *mockMessageStore) UnreadCount(recipientID string) (int, error) {
	count := 0
	for _, msg := range m.messages {
		if msg.To == recipientID && !msg.Read {
			count++
		}
	}
	return count, nil
}

type messageNotFoundError struct {
	id string
}

func (e *messageNotFoundError) Error() string {
	return "message not found: " + e.id
}

// Tests

func TestMessageService_Send(t *testing.T) {
	t.Run("sends message to inactive recipient", func(t *testing.T) {
		msgStore := newMockMessageStore()
		agentStore := newTestStore()
		tmux := &mockTmuxClient{sessions: make(map[string]bool)}

		svc := NewMessageService(msgStore, tmux, agentStore)

		msg, err := svc.Send("sender-001", "recipient-001", MessageTypeQuestion, "Test message", nil)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if msg.From != "sender-001" {
			t.Errorf("From = %q, want %q", msg.From, "sender-001")
		}
		if msg.To != "recipient-001" {
			t.Errorf("To = %q, want %q", msg.To, "recipient-001")
		}
		if msg.Type != MessageTypeQuestion {
			t.Errorf("Type = %q, want %q", msg.Type, MessageTypeQuestion)
		}
		if msg.Read {
			t.Error("message should not be marked as read for inactive recipient")
		}
	})

	t.Run("sends message to active recipient", func(t *testing.T) {
		msgStore := newMockMessageStore()
		agentStore := newTestStore()
		agentStore.Add(&Agent{ID: "recipient-001", Status: AgentStatusActive})
		tmux := &mockTmuxClient{sessions: map[string]bool{"recipient-001": true}}

		svc := NewMessageService(msgStore, tmux, agentStore)

		msg, err := svc.Send("sender-001", "recipient-001", MessageTypeQuestion, "Test message", nil)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !msg.Read {
			t.Error("message should be marked as read for active recipient")
		}
	})

	t.Run("messages to human are never auto-delivered", func(t *testing.T) {
		msgStore := newMockMessageStore()
		agentStore := newTestStore()
		tmux := &mockTmuxClient{sessions: make(map[string]bool)}

		svc := NewMessageService(msgStore, tmux, agentStore)

		msg, err := svc.Send("worker-001", HumanParticipantID, MessageTypeQuestion, "Need decision", nil)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if msg.Read {
			t.Error("messages to human should never be auto-marked as read")
		}
	})

	t.Run("rejects invalid message type", func(t *testing.T) {
		msgStore := newMockMessageStore()
		agentStore := newTestStore()
		tmux := &mockTmuxClient{sessions: make(map[string]bool)}

		svc := NewMessageService(msgStore, tmux, agentStore)

		_, err := svc.Send("sender", "recipient", MessageType("invalid"), "content", nil)

		if err == nil {
			t.Error("expected error for invalid message type")
		}
	})

	t.Run("includes related work reference", func(t *testing.T) {
		msgStore := newMockMessageStore()
		agentStore := newTestStore()
		tmux := &mockTmuxClient{sessions: make(map[string]bool)}

		svc := NewMessageService(msgStore, tmux, agentStore)

		relatedWork := "feature-oauth"
		msg, err := svc.Send("worker-001", "lead-001", MessageTypeCompletion, "Done", &relatedWork)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if msg.RelatedWork == nil || *msg.RelatedWork != "feature-oauth" {
			t.Error("related work not set correctly")
		}
	})
}

func TestMessageService_ListUnread(t *testing.T) {
	t.Run("returns only unread messages", func(t *testing.T) {
		msgStore := newMockMessageStore()
		msgStore.messages["msg-1"] = &Message{ID: "msg-1", To: "worker-001", Read: false}
		msgStore.messages["msg-2"] = &Message{ID: "msg-2", To: "worker-001", Read: true}
		msgStore.messages["msg-3"] = &Message{ID: "msg-3", To: "worker-001", Read: false}

		svc := NewMessageService(msgStore, nil, nil)

		msgs, err := svc.ListUnread("worker-001")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(msgs) != 2 {
			t.Errorf("got %d messages, want 2", len(msgs))
		}
	})
}

func TestMessageService_Read(t *testing.T) {
	t.Run("marks message as read", func(t *testing.T) {
		msgStore := newMockMessageStore()
		msgStore.messages["msg-1"] = &Message{ID: "msg-1", To: "worker-001", Read: false, Content: "Test"}

		svc := NewMessageService(msgStore, nil, nil)

		msg, err := svc.Read("msg-1")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if msg.Content != "Test" {
			t.Errorf("Content = %q, want %q", msg.Content, "Test")
		}
		// Check the store was updated
		if !msgStore.messages["msg-1"].Read {
			t.Error("message should be marked as read in store")
		}
	})
}

func TestMessageService_UnreadCount(t *testing.T) {
	t.Run("counts unread messages", func(t *testing.T) {
		msgStore := newMockMessageStore()
		msgStore.messages["msg-1"] = &Message{ID: "msg-1", To: "worker-001", Read: false}
		msgStore.messages["msg-2"] = &Message{ID: "msg-2", To: "worker-001", Read: true}
		msgStore.messages["msg-3"] = &Message{ID: "msg-3", To: "worker-001", Read: false}
		msgStore.messages["msg-4"] = &Message{ID: "msg-4", To: "other", Read: false}

		svc := NewMessageService(msgStore, nil, nil)

		count, err := svc.UnreadCount("worker-001")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if count != 2 {
			t.Errorf("count = %d, want 2", count)
		}
	})
}

func TestIsValidMessageType(t *testing.T) {
	validTypes := []string{"question", "answer", "assignment", "completion", "status", "info"}
	for _, typ := range validTypes {
		if !IsValidMessageType(typ) {
			t.Errorf("%q should be valid", typ)
		}
	}

	invalidTypes := []string{"invalid", "QUESTION", "Query", ""}
	for _, typ := range invalidTypes {
		if IsValidMessageType(typ) {
			t.Errorf("%q should be invalid", typ)
		}
	}
}
