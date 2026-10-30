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
	ChannelRepo repository.ChannelRepository
	MessageRepo repository.MessageRepository
	PlanRepo    repository.PlanRepository
	FriendRepo  repository.FriendRepository
	ConvRepo    repository.Conversation
	Config      *config.AppConfig

	// Services
	ChannelService     services.ChannelService
	FriendService      services.FriendService
	MessageService     services.MessageService
	ChatHistoryService services.ChatHistroyService
	PlanService        services.PlanService
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
	channelRepo := repository.NewChannelRepository(db)
	messageRepo := repository.NewMessageRepository(db)
	planRepo := repository.NewPlanRepository(db)
	friendRepo := repository.NewFriendRepository(db)
	convRepo := repository.NewConversation(db)
	sessionRepo := repository.NewSessionRepository(db)
	// services
	emailService := services.NewEmailService()
	authService := services.NewAuthService(userRepo, sessionRepo, emailService)

	//storage
	var fileStorage storage.FileStorage

	if cfg.MinIO.Endpoint == "" && cfg.MinIO.Username == "" && cfg.MinIO.Bucket == "" && cfg.MinIO.Password == "" {
	} else if cfg.MinIO.Endpoint == "" || cfg.MinIO.Username == "" || cfg.MinIO.Bucket == "" || cfg.MinIO.Password == "" {
		panic("incomplete MinIO configuration: all four MinIO environment variables must be set")
	} else {
		//storage
		minioStorage, err := storage.NewMinioStorage(*cfg)
		// this need to be remove later
		if err != nil {
			panic("failed to initialize MinIO: " + err.Error())
		}
		if err := minioStorage.EnsureBucket(context.Background()); err != nil {
			panic("MinIO connection failed: " + err.Error())
		}
		fileStorage = minioStorage
	}
	return &App{
		DB:     db,
		Hub:    hub,
		Config: cfg,

		//repository
		UserRepo:    userRepo,
		ChannelRepo: channelRepo,
		MessageRepo: messageRepo,
		PlanRepo:    planRepo,
		FriendRepo:  friendRepo,
		ConvRepo:    convRepo,

		//services
		ChatHistoryService: services.NewChatHistoryService(messageRepo),
		PlanService:        services.NewPlanService(planRepo, channelRepo, friendRepo, userRepo, hub),
		MessageService:     services.NewMessageService(messageRepo, userRepo, channelRepo),
		FriendService:      services.NewFriendService(userRepo, friendRepo),
		ChannelService:     services.NewChannelService(db, channelRepo, userRepo),
		EmailService:       emailService,
		AuthService:        authService,

		//storage
		FileStorage: fileStorage,
	}
}
