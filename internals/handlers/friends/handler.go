package friends

import "chat-server/internals/services"

type FriendHandler struct {
	friendService services.FriendService
}

func NewFriendHandler(friendService services.FriendService) *FriendHandler {
	return &FriendHandler{
		friendService: friendService,
	}
}
