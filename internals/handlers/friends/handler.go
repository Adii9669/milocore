package friends

import (
	"chat-server/internals/repository"
)

type FriendHandler struct {
	FrndRepo repository.FriendRepository
	UserRepo repository.UserRepository
}

func NewFriendHandler(repo repository.FriendRepository, userRepo repository.UserRepository) *FriendHandler {
	return &FriendHandler{
		FrndRepo: repo,
		UserRepo: userRepo,
	}
}
