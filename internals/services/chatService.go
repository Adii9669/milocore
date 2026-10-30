package services

import (
	"chat-server/internals/db/models"
	"chat-server/internals/repository"
	"chat-server/internals/transport/dto"
	"chat-server/internals/transport/mapper"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type ChatHistroyService interface {
	ChannelHistory(
		ctx context.Context,
		channelID uuid.UUID,
		currentUser uuid.UUID,
		limit int,
		cursor *time.Time,
	) ([]dto.MessageResponse, error)

	DmHistory(
		ctx context.Context,
		userA uuid.UUID,
		userB uuid.UUID,
		limit int,
		cursor *time.Time,
	) ([]dto.MessageResponse, error)
}

type chathistroyservice struct {
	messageRepo repository.MessageRepository
}

func NewChatHistoryService(repo repository.MessageRepository) ChatHistroyService {
	return &chathistroyservice{
		messageRepo: repo,
	}
}

func (s *chathistroyservice) ChannelHistory(
	ctx context.Context,
	channelID uuid.UUID,
	currentUser uuid.UUID,
	limit int,
	cursor *time.Time,
) ([]dto.MessageResponse, error) {

	if limit <= 0 || limit >= 100 {
		limit = 50
	}
	var messages []models.Message
	var err error
	if deletionRepo, ok := s.messageRepo.(repository.MessageDeletionRepository); ok {
		messages, err = deletionRepo.GetChannelMessageHistoryForUser(ctx, channelID, currentUser, limit, cursor)
	} else {
		messages, err = s.messageRepo.GetChannelMessageHistory(ctx, channelID, limit, cursor)
	}
	if err != nil {
		return nil, errors.New("Failed to get channel Histroy")
	}

	var responses []dto.MessageResponse

	for _, msg := range messages {
		resp := mapper.ToChannelMessageResponse(&msg, currentUser)
		responses = append(responses, resp)
	}
	return responses, nil
}

func (s *chathistroyservice) DmHistory(
	ctx context.Context,
	userA uuid.UUID,
	userB uuid.UUID,
	limit int,
	cursor *time.Time,
) ([]dto.MessageResponse, error) {

	if userA == uuid.Nil || userB == uuid.Nil {
		return nil, errors.New("invalid users")
	}

	if userA == userB {
		return nil, errors.New("cannot dm yourself")
	}

	if limit <= 0 || limit > 100 {
		limit = 50
	}

	var messages []models.Message
	var err error
	if deletionRepo, ok := s.messageRepo.(repository.MessageDeletionRepository); ok {
		messages, err = deletionRepo.GetDmMessageHistoryForUser(ctx, userA, userB, userA, limit, cursor)
	} else {
		messages, err = s.messageRepo.GetDmMessageHistory(ctx, userA, userB, limit, cursor)
	}
	if err != nil {
		return nil, errors.New("Failed to get Histroy")
	}

	//creating a slice to change all the models into dto
	var responses []dto.MessageResponse

	for _, msg := range messages {
		resp := mapper.ToDMMessageResponse(&msg, userA)
		responses = append(responses, resp)
	}

	return responses, nil
}
