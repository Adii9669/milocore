package requests

import "encoding/json"

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
