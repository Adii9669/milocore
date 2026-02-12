package chathistory

import "chat-server/internals/services"

type Handler struct {
	service services.ChatHistroyService
}

func NewHandler(service services.ChatHistroyService) *Handler {
	return &Handler{
		service: service,
	}
}
