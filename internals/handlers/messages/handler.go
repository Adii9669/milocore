package messages

import "chat-server/internals/repository"

type MessageHandler struct {
	MessageRepo repository.MessageRepository
}

func NewMessageHandler(repo repository.MessageRepository) *MessageHandler {
	return &MessageHandler{
		MessageRepo: repo,
	}
}
