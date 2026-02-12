package websockets

import "time"

type IncomingMessage struct {
	Type       string `json:"type"`        // "dm" | "crew"
	ReceiverID string `json:"receiver_id"` // userID or crewID
	Content    string `json:"content"`
}

type OutgoingMessage struct {
	ID         string    `json:"id"`
	Type       string    `json:"type"`
	SenderID   string    `json:"sender_id"`
	ReceiverID string    `json:"receiver_id"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
}

type AckMessage struct {
	Type      string `json:"type"` // "ack"
	MessageID string `json:"message_id"`
	Status    string `json:"status"` // "sent" | "delivered"
}
