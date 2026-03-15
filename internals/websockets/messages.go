package websockets

type MessageType string

const (
	TextMessage MessageType = "text_message"
	FileMessage MessageType = "file_message"
	TypingEvent MessageType = "typing"
)

// Creating the protocol for sending the message through the websockets
type WSMessage struct {
	Type MessageType `json:"type"`

	CrewID     string `json:"crewId,omitempty"`
	ReceiverID string `json:"receiverId,omitempty"`

	Content string `json:"content,omitempty"`
}

type AckMessage struct {
	Type      string `json:"type"` // "ack"
	MessageID string `json:"message_id"`
	Status    string `json:"status"` // "sent" | "delivered"
}
