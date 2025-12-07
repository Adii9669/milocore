package app

import (
	"chat-server/internals/config"
	"chat-server/internals/repository"
	"chat-server/internals/websockets"

	"gorm.io/gorm"
)

type App struct {
	DB  *gorm.DB
	Hub *websockets.Hub

	UserRepo    repository.UserRepository
	CrewRepo    repository.CrewRepository
	MessageRepo repository.MessageRepository
	FriendRepo  repository.FriendRepository
	Config      *config.AppConfig
}

func NewApp(
	db *gorm.DB,
	hub *websockets.Hub,
	cfg *config.AppConfig,
) *App {

	return &App{
		DB:     db,
		Hub:    hub,
		Config: cfg,

		UserRepo:    repository.NewUserRepository(),
		CrewRepo:    repository.NewCrewRepository(),
		MessageRepo: repository.NewMessageRepository(),
		FriendRepo:  repository.NewFriendRepository(),
	}
}
