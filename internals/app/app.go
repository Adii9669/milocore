package app

import (
	"chat-server/internals/config"
	"chat-server/internals/repository"
	"chat-server/internals/services"
	"chat-server/internals/websockets"
	"chat-server/storage"
	"context"

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
	EmailService       *services.EmailService
	AuthService        *services.AuthService

	//Storage
	FileStorage storage.FileStorage
}

// Constructor for the app
func NewApp(
	db *gorm.DB,
	hub *websockets.Hub,
	cfg *config.AppConfig,
) *App {
	userRepo := repository.NewUserRepository(db)
	crewRepo := repository.NewCrewRepository(db)
	messageRepo := repository.NewMessageRepository(db)
	friendRepo := repository.NewFriendRepository(db)
	convRepo := repository.NewConversation(db)
	sessionRepo := repository.NewSessionRepository(db)
	// services
	emailService := services.NewEmailService()
	authService := services.NewAuthService(userRepo, sessionRepo, emailService)

	//storage
	minioStorage, err := storage.NewMinioStorage(*cfg)
	// this need to be remove later
	if err != nil {
		panic("failed to initialize MinIO: " + err.Error())
	}
	if err := minioStorage.CheckConnection(context.Background()); err != nil {
		panic("MinIO connection failed: " + err.Error())
	}

	return &App{
		DB:     db,
		Hub:    hub,
		Config: cfg,

		//repository
		UserRepo:    userRepo,
		CrewRepo:    crewRepo,
		MessageRepo: messageRepo,
		FriendRepo:  friendRepo,
		ConvRepo:    convRepo,

		//services
		ChatHistoryService: services.NewChatHistoryService(messageRepo),
		MessageService:     services.NewMessageService(messageRepo, userRepo, crewRepo),
		FriendService:      services.NewFriendService(userRepo, friendRepo),
		CrewService:        services.NewCrewService(db, crewRepo, userRepo),
		EmailService:       emailService,
		AuthService:        authService,

		//storage
		FileStorage: minioStorage,
	}
}
