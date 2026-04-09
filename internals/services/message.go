package services

import (
	"chat-server/internals/db/models"
	"chat-server/internals/events"
	"chat-server/internals/repository"
	"context"
	"errors"
	"log"

	"github.com/google/uuid"
)

type IncomingMessage struct {
	Type       string     `json:"type"` // "message"
	Content    string     `json:"content"`
	CrewID     *uuid.UUID `json:"crewId"`
	ReceiverID *uuid.UUID `json:"receiverId"`
}

type MessageService interface {
	HandleIncomingMessage(
		ctx context.Context,
		senderID uuid.UUID,
		payload IncomingMessage) (events.MessageEvent, error)
	MarkDelivered(
		ctx context.Context,
		msgID uuid.UUID,
	) error
	MarkRead(
		ctx context.Context,
		userID, OtherUserID uuid.UUID,
	) error
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
	senderID uuid.UUID,
	payload IncomingMessage) (events.MessageEvent, error) {

	//check the payload first
	if payload.Type != "dm" && payload.Type != "crew" {
		return events.MessageEvent{}, errors.New("invalid message type")
	}

	if payload.Type == "dm" {
		if payload.ReceiverID == nil || payload.CrewID != nil {
			return events.MessageEvent{}, errors.New("invalid message type")
		}
	}

	if payload.Type == "crew" {
		if payload.ReceiverID != nil || payload.CrewID == nil {
			return events.MessageEvent{}, errors.New("invalid message type")
		}
	}

	if payload.Content == "" {
		return events.MessageEvent{}, errors.New("empty message")
	}

	if payload.CrewID == nil && payload.ReceiverID == nil {
		return events.MessageEvent{}, errors.New("no target specified")
	}

	if payload.CrewID != nil && payload.ReceiverID != nil {
		return events.MessageEvent{}, errors.New("ambiguous target")
	}

	//check the userID exist or not ?

	//checking the senderID
	sender, err := s.userRepo.FindByID(ctx, senderID)
	if err != nil {
		return events.MessageEvent{}, err
	}

	// =========================
	// DM MESSAGE
	// =========================
	if payload.ReceiverID != nil {
		//reciverID check (validate)
		receiver := *payload.ReceiverID

		//if exist (check if exist or not )
		exists, err := s.userRepo.ExistByID(ctx, receiver)
		if err != nil {
			return events.MessageEvent{}, err
		}
		if !exists {
			return events.MessageEvent{}, errors.New("receiver does not exist")
		}

		//msg store the raw db models  for the message
		log.Printf("Testesing receiver %v", *payload.ReceiverID)
		msg := &models.Message{
			ID:         uuid.New(),
			SenderID:   senderID,
			ReceiverID: payload.ReceiverID,
			Content:    &payload.Content,
		}

		//this is the mapper converting the database models into the dto for sending back the response

		msg.Sender = *sender

		if err := s.messageRepo.SaveMessage(ctx, msg); err != nil {
			return events.MessageEvent{}, err
		}
		// resp := mapper.ToDMMessageResponse(msg, senderID)
		events := events.MessageEvent{
			ID:         msg.ID,
			Type:       "dm",
			Content:    *msg.Content,
			CreatedAt:  msg.CreatedAt,
			SenderID:   sender.ID,
			SenderName: sender.Name,
			ReceiverID: payload.ReceiverID,
			CrewID:     nil,
		}

		// return &MessageResult{
		// 	Response:   &resp,
		// 	SenderID:   senderID,
		// 	ReceiverID: payload.ReceiverID,
		// 	CrewID:     nil,
		// }, nil
		return events, nil
	}

	// =========================
	//  CREW MESSAGE
	// =========================
	crewUUID := *payload.CrewID

	//if exist (check if exist or not )
	exists, err := s.crewRepo.ExistByID(ctx, crewUUID)
	if err != nil {
		return events.MessageEvent{}, err
	}
	if !exists {
		return events.MessageEvent{}, errors.New("receiver does not exist")
	}

	// check if a member of the crew or not ?
	isMember, err := s.crewRepo.IsMember(ctx, crewUUID, senderID)
	if err != nil {
		return events.MessageEvent{}, err
	}
	if !isMember {
		return events.MessageEvent{}, errors.New("not a member of this crew")
	}

	msg := &models.Message{
		ID:       uuid.New(),
		CrewID:   payload.CrewID,
		SenderID: senderID,
		Content:  &payload.Content,
	}

	if err := s.messageRepo.SaveMessage(ctx, msg); err != nil {
		return events.MessageEvent{}, err
	}
	msg.Sender = *sender
	events := events.MessageEvent{
		ID:         msg.ID,
		Type:       "dm",
		Content:    *msg.Content,
		CreatedAt:  msg.CreatedAt,
		SenderID:   sender.ID,
		SenderName: sender.Name,
		ReceiverID: nil,
		CrewID:     payload.CrewID,
	}

	return events, nil
	// resp := mapper.ToCrewMessageResponse(msg, senderID)
	// return &MessageResult{
	// 	Response:   &resp,
	// 	SenderID:   senderID,
	// 	ReceiverID: nil,
	// 	CrewID:     payload.CrewID,
	// }, nil
}

func (s *messageService) MarkDelivered(ctx context.Context, messageID uuid.UUID) error {
	return s.messageRepo.MarkDelivered(ctx, messageID)
}

func (s *messageService) MarkRead(ctx context.Context, userID, OtherUserID uuid.UUID) error {
	return s.messageRepo.MarkRead(ctx, userID, OtherUserID)
}
