package app

import (
	"chat-server/internals/config"
	"chat-server/internals/repository"
	"chat-server/internals/websockets"

	"gorm.io/gorm"
)

// The App struct like a contract to create the app
type App struct {
	//Database and websockets(Hub)
	DB  *gorm.DB
	Hub *websockets.Hub

	//Repositories for the app
	UserRepo    repository.UserRepository
	CrewRepo    repository.CrewRepository
	MessageRepo repository.MessageRepository
	FriendRepo  repository.FriendRepository
	ConvRepo    repository.Conversation
	Config      *config.AppConfig
}

// Constructor for the app
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
		ConvRepo:    repository.NewConversation(),
	}
}
