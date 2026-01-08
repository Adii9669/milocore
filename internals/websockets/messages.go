package websockets

type IncomingMessage struct {
	Type       string `json:"type"`
	SenderID   string `json:"sender_id"`
	ReceiverID string `json:"reciever_id"`
	Content    string `json:"content"`
}

type OutgoingMessage struct {
	Type       string `json:"type"`
	SenderID   string `json:"sender_id"`
	ReceiverID string `json:"reciever_id"`
	Content    string `json:"content"`
}
