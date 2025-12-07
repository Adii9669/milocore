package auth

import "chat-server/internals/repository"

type AuthHandler struct {
	UserRepo repository.UserRepository
}

func NewAuthHandler(userRepo repository.UserRepository) *AuthHandler {
	return &AuthHandler{UserRepo: userRepo}
}
