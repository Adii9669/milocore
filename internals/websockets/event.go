package websockets

import (
	"time"

	"github.com/google/uuid"
)

type WSEvent struct {
	Event string `json:"event"`
	Data  any    `json:"data"`
}

type EventType string

const (
	MessageSend      EventType = "message.send"
	MessageDelivered EventType = "message.delivered"
	MessageRead      EventType = "message.read"
)

type WSMessagePayload struct {
	ID         string     `json:"id"`
	Type       string     `json:"type"`
	EventType  EventType  `json:"event"`
	Content    string     `json:"content"`
	CreatedAt  time.Time  `json:"createdAt"`
	Sender     SenderDTO  `json:"sender"`
	ReceiverID *uuid.UUID `json:"receiverId,omitempty"`
	CrewID     *uuid.UUID `json:"crewId,omitempty"`
}

type SenderDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type WSAck struct {
	MessageID uuid.UUID `json:"messageId"`
	Status    string    `json:"status"`
}
