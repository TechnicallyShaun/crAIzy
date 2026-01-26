package domain

import (
	"time"

	"github.com/google/uuid"
)

// MessageType represents the type/intent of a message.
type MessageType string

const (
	MessageTypeQuestion   MessageType = "question"   // Needs answer before proceeding
	MessageTypeAnswer     MessageType = "answer"     // Response to a question
	MessageTypeAssignment MessageType = "assignment" // Work being delegated
	MessageTypeCompletion MessageType = "completion" // Task/work finished
	MessageTypeStatus     MessageType = "status"     // Progress update
	MessageTypeInfo       MessageType = "info"       // General information
)

// ValidMessageTypes contains all valid message types.
var ValidMessageTypes = []MessageType{
	MessageTypeQuestion,
	MessageTypeAnswer,
	MessageTypeAssignment,
	MessageTypeCompletion,
	MessageTypeStatus,
	MessageTypeInfo,
}

// IsValidMessageType checks if a string is a valid message type.
func IsValidMessageType(t string) bool {
	for _, valid := range ValidMessageTypes {
		if string(valid) == t {
			return true
		}
	}
	return false
}

// Message represents a message between agents or between agents and humans.
type Message struct {
	ID          string      // Unique identifier (UUID)
	From        string      // Sender ID (tmux session name or "human")
	To          string      // Recipient ID (tmux session name or "human")
	Type        MessageType // Message type/intent
	Content     string      // Message content
	RelatedWork *string     // Optional work item reference
	Read        bool        // Whether the message has been read
	CreatedAt   time.Time   // When the message was sent
	ReadAt      *time.Time  // When the message was read (nil if unread)
}

// NewMessage creates a new message with a generated UUID.
func NewMessage(from, to string, msgType MessageType, content string, relatedWork *string) *Message {
	return &Message{
		ID:          uuid.New().String(),
		From:        from,
		To:          to,
		Type:        msgType,
		Content:     content,
		RelatedWork: relatedWork,
		Read:        false,
		CreatedAt:   time.Now(),
	}
}

// HumanParticipantID is the reserved ID for human participants.
const HumanParticipantID = "human"
