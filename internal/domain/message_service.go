package domain

import (
	"fmt"

	"github.com/TechnicallyShaun/crAIzy/internal/logging"
)

// MessageService handles message operations.
type MessageService struct {
	store  IMessageStore
	tmux   ITmuxClient
	agents IAgentStore
}

// NewMessageService creates a new MessageService with the given dependencies.
func NewMessageService(store IMessageStore, tmux ITmuxClient, agents IAgentStore) *MessageService {
	return &MessageService{
		store:  store,
		tmux:   tmux,
		agents: agents,
	}
}

// Send creates and delivers a message.
// If the recipient is active (has a tmux session), the message is delivered immediately.
// Otherwise, it is queued for delivery on startup.
func (s *MessageService) Send(from, to string, msgType MessageType, content string, relatedWork *string) (*Message, error) {
	logging.Entry("from", from, "to", to, "type", msgType)

	if !IsValidMessageType(string(msgType)) {
		err := fmt.Errorf("invalid message type: %s", msgType)
		logging.Error(err, "type", msgType)
		return nil, err
	}

	msg := NewMessage(from, to, msgType, content, relatedWork)

	// 1. Persist to DB
	if err := s.store.Save(msg); err != nil {
		logging.Error(err, "msgID", msg.ID)
		return nil, fmt.Errorf("failed to save message: %w", err)
	}

	// 2. If recipient is active, deliver immediately
	if s.isActive(to) {
		s.deliverToTmux(msg)
		if err := s.store.MarkRead(msg.ID); err != nil {
			// Log but don't fail - message is saved
			logging.Error(err, "msgID", msg.ID, "action", "mark read after delivery")
		}
		msg.Read = true
	}

	logging.Info("message sent, msgID=%s, from=%s, to=%s", msg.ID, from, to)
	return msg, nil
}

// ListUnread returns all unread messages for a recipient.
func (s *MessageService) ListUnread(recipientID string) ([]*Message, error) {
	logging.Entry("recipientID", recipientID)
	return s.store.ListUnread(recipientID)
}

// List returns messages for a recipient with a limit (0 = no limit).
func (s *MessageService) List(recipientID string, limit int) ([]*Message, error) {
	logging.Entry("recipientID", recipientID, "limit", limit)
	return s.store.List(recipientID, limit)
}

// Read retrieves a message and marks it as read.
func (s *MessageService) Read(messageID string) (*Message, error) {
	logging.Entry("messageID", messageID)

	msg, err := s.store.Get(messageID)
	if err != nil {
		logging.Error(err, "messageID", messageID)
		return nil, err
	}

	if !msg.Read {
		if err := s.store.MarkRead(messageID); err != nil {
			logging.Error(err, "messageID", messageID, "action", "mark read")
			return nil, fmt.Errorf("failed to mark message as read: %w", err)
		}
	}

	return msg, nil
}

// UnreadCount returns the count of unread messages for a recipient.
func (s *MessageService) UnreadCount(recipientID string) (int, error) {
	logging.Entry("recipientID", recipientID)
	return s.store.UnreadCount(recipientID)
}

// MarkRead marks a message as read.
// This is exposed for startup delivery in AgentService.
func (s *MessageService) MarkRead(messageID string) error {
	logging.Entry("messageID", messageID)
	return s.store.MarkRead(messageID)
}

// isActive checks if a recipient is active (has a running tmux session).
func (s *MessageService) isActive(agentID string) bool {
	// Human messages are never auto-delivered
	if agentID == HumanParticipantID {
		return false
	}

	agent := s.agents.Get(agentID)
	if agent == nil {
		return false
	}

	return s.tmux.SessionExists(agent.ID)
}

// deliverToTmux sends a notification to the recipient's tmux session.
func (s *MessageService) deliverToTmux(msg *Message) {
	agent := s.agents.Get(msg.To)
	if agent == nil {
		return
	}

	notification := fmt.Sprintf("\n[MESSAGE from %s (%s)]: %s\n",
		msg.From, msg.Type, msg.Content)

	if err := s.tmux.SendKeys(agent.ID, notification); err != nil {
		logging.Error(err, "agentID", agent.ID, "action", "deliver to tmux")
	}
}
