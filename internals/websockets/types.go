package websockets

import "github.com/google/uuid"

type MessageType string
type MessageStatus string

const (
	TextMessage     MessageType   = "text_message"
	FileMessage     MessageType   = "file_message"
	TypingEvent     MessageType   = "typing"
	TypeDM          MessageType   = "dm"
	TypeCrewMessage MessageType   = "crew_message"
	TypeTyping      MessageType   = "typing"
	TypeDelivered   MessageType   = "message.delivered"
	TypeRead        MessageType   = "message.read"
	StatusSent      MessageStatus = "sent"
	StatusDelivered MessageStatus = "delivered"
	StatusRead      MessageStatus = "read"
)

// Creating the protocol for sending the message through the websockets
type WSMessage struct {
	Type       MessageType        `json:"type"`
	CrewID     uuid.UUID          `json:"crewId,omitempty"`
	ReceiverID uuid.UUID          `json:"receiverId,omitempty"`
	UserID     uuid.UUID          `json:"user_id,omitempty"`
	Content    string             `json:"content,omitempty"`
	MessageID  *uuid.UUID         `json:"messageId,omitempty"`
	Status     MessageStatusEvent `json:"status"`
	FileURL    *string
	FileName   *string
	FileSize   *int64
	MimeType   *string
}

type MessageStatusEvent struct {
	Type      string `json:"type"` // "ack"
	MessageID string `json:"message_id"`
	Status    string `json:"status"` // "sent" | "delivered"
}
