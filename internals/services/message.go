package services

import (
	"chat-server/internals/db/models"
	"chat-server/internals/repository"
	"chat-server/internals/transport/mapper"
	"context"
	"errors"

	"github.com/google/uuid"
)

type IncomingMessage struct {
	Type       string  `json:"type"` // "message"
	Content    string  `json:"content"`
	CrewID     *string `json:"crewId"`
	ReceiverID *string `json:"receiverId"`
}

type MessageService interface {
	HandleIncomingMessage(
		ctx context.Context,
		senderID string,
		payload IncomingMessage) (*MessageResult, error)
}

type messageService struct {
	messageRepo repository.MessageRepository
	userRepo    repository.UserRepository
	crewRepo    repository.CrewRepository
}

func NewMessageService(
	messageRepo repository.MessageRepository,
	userRepo repository.UserRepository,
	crewRepo repository.CrewRepository,
) MessageService {
	return &messageService{
		messageRepo: messageRepo,
		userRepo:    userRepo,
		crewRepo:    crewRepo,
	}
}

func (s *messageService) HandleIncomingMessage(
	ctx context.Context,
	senderID string,
	payload IncomingMessage) (*MessageResult, error) {

	//check the payload first
	if payload.Type != "dm" && payload.Type != "crew" {
		return nil, errors.New("invalid message type")
	}
	if payload.Type == "dm" {
		if payload.ReceiverID == nil || payload.CrewID != nil {
			return nil, errors.New("invalid message type")
		}
	}

	if payload.Type == "crew" {
		if payload.ReceiverID != nil || payload.CrewID == nil {
			return nil, errors.New("invalid message type")
		}
	}

	if payload.Content == "" {
		return nil, errors.New("empty message")
	}

	if payload.CrewID == nil && payload.ReceiverID == nil {
		return nil, errors.New("no target specified")
	}

	if payload.CrewID != nil && payload.ReceiverID != nil {
		return nil, errors.New("ambiguous target")
	}

	//check the userID exist or not ?
	senderUUID, err := uuid.Parse(senderID)
	if err != nil {
		return nil, errors.New("invalid sender Id")
	}

	//checking the senderID
	sender, err := s.userRepo.FindByID(ctx, senderUUID)
	if err != nil {
		return nil, err
	}

	// =========================
	// DM MESSAGE
	// =========================
	if payload.ReceiverID != nil {
		//reciverID check (validate)
		receiver, err := uuid.Parse(*payload.ReceiverID)
		if err != nil {
			return nil, errors.New("invalid receiver ID")
		}

		//if exist (check if exist or not )

		exists, err := s.userRepo.ExistByID(ctx, receiver)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, errors.New("receiver does not exist")
		}

		//msg store the raw db models  for the message
		msg := &models.Message{
			ID:         uuid.New(),
			SenderID:   senderID,
			ReceiverID: payload.ReceiverID,
			Content:    &payload.Content,
		}

		if err := s.messageRepo.SaveMessage(ctx, msg); err != nil {
			return nil, err
		}

		//this is the mapper converting the database models into the dto for sending back the response
		resp := mapper.ToDMMessageResponse(msg, senderID)
		return &MessageResult{
			Response: &resp,
		}, nil
	}

	// =========================
	//  CREW MESSAGE
	// =========================
	crewUUID, err := uuid.Parse(*payload.CrewID)
	if err != nil {
		return nil, errors.New("invalid crew ID")
	}

	//if exist (check if exist or not )
	exists, err := s.crewRepo.ExistByID(ctx, crewUUID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("receiver does not exist")
	}

	// check if a member of the crew or not ?
	isMember, err := s.crewRepo.IsMember(ctx, crewUUID, senderUUID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, errors.New("not a member of this crew")
	}

	msg := &models.Message{
		ID:       uuid.New(),
		CrewID:   payload.CrewID,
		SenderID: senderID,
		Content:  &payload.Content,
	}

	if err := s.messageRepo.SaveMessage(ctx, msg); err != nil {
		return nil, err
	}
	msg.Sender = *sender
	resp := mapper.ToCrewMessageResponse(msg, senderID)
	return &MessageResult{
		Response:   &resp,
		SenderID:   senderID,
		ReceiverID: nil,
		CrewID:     payload.CrewID,
	}, nil
}
