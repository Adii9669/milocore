package websockets

import (
	"time"

	"github.com/google/uuid"
)

type WSEvent struct {
	Event string `json:"event"`
	Data  any    `json:"data"`
}

type WSMessagePayload struct {
	ID         string    `json:"id"`
	Type       string    `json:"type"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"createdAt"`
	Sender     SenderDTO `json:"sender"`
	ReceiverID *string   `json:"receiverId,omitempty"`
	CrewID     *string   `json:"crewId,omitempty"`
}

type SenderDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type WSAck struct {
	MessageID uuid.UUID `json:"messageId"`
	Status    string    `json:"status"`
}
