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
	MessageDeleted   EventType = "message.deleted"
)

type PlanEventData struct {
	EventID      string     `json:"eventId"`
	PlanID       string     `json:"planId"`
	ChannelID    *uuid.UUID `json:"channelId,omitempty"`
	DMUserLowID  *uuid.UUID `json:"dmUserLowId,omitempty"`
	DMUserHighID *uuid.UUID `json:"dmUserHighId,omitempty"`
}

type WSMessagePayload struct {
	ID             string     `json:"id"`
	Type           string     `json:"type"`
	EventType      EventType  `json:"event"`
	Content        string     `json:"content"`
	CreatedAt      time.Time  `json:"createdAt"`
	Sender         SenderDTO  `json:"sender"`
	ReceiverID     *uuid.UUID `json:"receiverId,omitempty"`
	ChannelID      *uuid.UUID `json:"channelId,omitempty"`
	FileURL        *string    `json:"fileUrl,omitempty"`
	FilePreviewURL *string    `json:"filePreviewUrl,omitempty"`
	FileName       *string    `json:"fileName,omitempty"`
	FileSize       *int64     `json:"fileSize,omitempty"`
	MimeType       *string    `json:"mimeType,omitempty"`
}

type SenderDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type WSAck struct {
	MessageID uuid.UUID `json:"messageId"`
	Status    string    `json:"status"`
}

type WSMessageDeletionPayload struct {
	MessageID  string     `json:"messageId"`
	SenderID   string     `json:"senderId"`
	ReceiverID *uuid.UUID `json:"receiverId,omitempty"`
	ChannelID  *uuid.UUID `json:"channelId,omitempty"`
}
