package requests

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type IncomingEnvelope struct {
	Action  string          `json:"action"` // e.g., "message"
	Payload json.RawMessage `json:"payload"`
}

// Payload for message action
type IncomingPayloadMessage struct {
	Type        string  `json:"type"` // "dm" or "crew" (we handle "dm" now)
	ToUserID    *string `json:"to_user_id,omitempty"`
	Content     string  `json:"content"`
	ClientNonce *string `json:"client_nonce,omitempty"`
}

type MessageResponse struct {
	ID         uuid.UUID `json:"ID"`
	SenderID   string    `json:"SenderID"`
	SenderName string    `json:"SenderName"` // Add this for frontend
	ReceiverID *string   `json:"ReceiverID,omitempty"`
	CrewID     *string   `json:"CrewID,omitempty"`
	Content    *string   `json:"Content"`
	CreatedAt  time.Time `json:"CreatedAt"`
}
