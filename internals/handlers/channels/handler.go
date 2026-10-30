package channels

import (
	"chat-server/internals/services"
)

type ChannelHandler struct {
	channelService services.ChannelService
}

func NewChannelHandler(channelService services.ChannelService) *ChannelHandler {
	return &ChannelHandler{
		channelService: channelService,
	}
}
