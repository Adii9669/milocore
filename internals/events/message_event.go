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

	ReceiverID *uuid.UUID
	CrewID     *uuid.UUID
}
