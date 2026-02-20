package services

import (
	"chat-server/internals/repository"
	"chat-server/internals/transport/dto"
	"chat-server/internals/transport/mapper"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type ChatHistroyService interface {
	CrewHistory(
		ctx context.Context,
		crewID uuid.UUID,
		currentUser string,
		limit int,
		cursor *time.Time,
	) ([]dto.MessageResponse, error)

	DmHistory(
		ctx context.Context,
		userA string,
		userB string,
		limit int,
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

func (s *chathistroyservice) CrewHistory(
	ctx context.Context,
	crewID uuid.UUID,
	currentUser string,
	limit int,
	cursor *time.Time,
) ([]dto.MessageResponse, error) {

	if limit <= 0 || limit >= 100 {
		limit = 50
	}
	messages, err := s.messageRepo.GetCrewMessageHistory(ctx, crewID, limit, cursor)
	if err != nil {
		return nil, errors.New("Failed to get crew Histroy")
	}

	var responses []dto.MessageResponse

	for _, msg := range messages {
		resp := mapper.ToCrewMessageResponse(&msg, currentUser)
		responses = append(responses, resp)
	}
	return responses, nil
}

func (s *chathistroyservice) DmHistory(ctx context.Context, userA string, userB string, limit int) ([]dto.MessageResponse, error) {
	if userA == "" || userB == "" {
		return nil, errors.New("invalid users")
	}

	if userA == userB {
		return nil, errors.New("cannot dm yourself")
	}

	if limit <= 0 || limit > 100 {
		limit = 50
	}

	messages, err := s.messageRepo.GetDmMessageHistory(ctx, userA, userB, limit)
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
