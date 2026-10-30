package events

import (
	"time"

	"github.com/google/uuid"
)

type EventType string

const (
	MessageSend      EventType = "message.send"
	MessageDelivered EventType = "message.delivered"
	MessageRead      EventType = "message.read"
)

type MessageEvent struct {
	ID        uuid.UUID
	Type      string
	EventType EventType
	Content   string
	CreatedAt time.Time

	SenderID   uuid.UUID
	SenderName string

	ReceiverID     *uuid.UUID
	ChannelID      *uuid.UUID
	FileURL        *string
	FilePreviewURL *string
	FileName       *string
	FileSize       *int64
	MimeType       *string
}

type MessageDeletionEvent struct {
	MessageID  uuid.UUID
	SenderID   uuid.UUID
	ReceiverID *uuid.UUID
	ChannelID  *uuid.UUID
}
