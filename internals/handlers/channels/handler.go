package channels

import (
	"chat-server/internals/services"
	"chat-server/internals/websockets"
)

type ChannelHandler struct {
	channelService services.ChannelService
	hub            *websockets.Hub
}

func NewChannelHandler(channelService services.ChannelService, hub *websockets.Hub) *ChannelHandler {
	return &ChannelHandler{
		channelService: channelService,
		hub:            hub,
	}
}
