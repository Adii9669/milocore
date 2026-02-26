package app

import (
	"chat-server/internals/config"
	"chat-server/internals/repository"
	"chat-server/internals/services"
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

	// Services
	CrewService        services.CrewService
	FriendService      services.FriendService
	MessageService     services.MessageService
	ChatHistoryService services.ChatHistroyService
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

		UserRepo:           repository.NewUserRepository(db),
		CrewRepo:           repository.NewCrewRepository(db),
		MessageRepo:        repository.NewMessageRepository(db),
		FriendRepo:         repository.NewFriendRepository(db),
		ConvRepo:           repository.NewConversation(db),
		ChatHistoryService: services.NewChatHistoryService(repository.NewMessageRepository(db)),
		MessageService: services.NewMessageService(
			repository.NewMessageRepository(db),
			repository.NewUserRepository(db),
			repository.NewCrewRepository(db)),
		FriendService: services.NewFriendService(
			repository.NewUserRepository(db),
			repository.NewFriendRepository(db),
		),
		CrewService: services.NewCrewService(
			db,
			repository.NewCrewRepository(db),
			repository.NewUserRepository(db),
		),
	}
}
