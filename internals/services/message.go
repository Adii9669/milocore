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

type AttachmentInput struct {
	ObjectKey string
	FileName  string
	FileSize  int64
	MimeType  string
}

type IncomingMessage struct {
	Type       string     `json:"type"` // "message"
	Content    string     `json:"content"`
	ChannelID  *uuid.UUID `json:"channelId"`
	ReceiverID *uuid.UUID `json:"receiverId"`
}

type MessageService interface {
	HandleIncomingMessage(
		ctx context.Context,
		senderID uuid.UUID,
		payload IncomingMessage) (events.MessageEvent, error)
	CreateAttachmentMessage(ctx context.Context, senderID uuid.UUID, target IncomingMessage, file AttachmentInput) (events.MessageEvent, error)
	GetAttachment(ctx context.Context, requesterID, messageID uuid.UUID) (*models.Message, error)
	DeleteMessage(ctx context.Context, requesterID, messageID uuid.UUID, scope string) (events.MessageDeletionEvent, *string, error)
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
	channelRepo repository.ChannelRepository
}

func NewMessageService(
	messageRepo repository.MessageRepository,
	userRepo repository.UserRepository,
	channelRepo repository.ChannelRepository,

) MessageService {
	return &messageService{
		messageRepo: messageRepo,
		userRepo:    userRepo,
		channelRepo: channelRepo,
	}
}

func (s *messageService) HandleIncomingMessage(
	ctx context.Context,
	senderID uuid.UUID,
	payload IncomingMessage) (events.MessageEvent, error) {

	//check the payload first
	if payload.Type != "dm" && payload.Type != "channel" {
		return events.MessageEvent{}, errors.New("invalid message type")
	}

	if payload.Type == "dm" {
		if payload.ReceiverID == nil || payload.ChannelID != nil {
			return events.MessageEvent{}, errors.New("invalid message type")
		}
	}

	if payload.Type == "channel" {
		if payload.ReceiverID != nil || payload.ChannelID == nil {
			return events.MessageEvent{}, errors.New("invalid message type")
		}
	}

	if payload.Content == "" {
		return events.MessageEvent{}, errors.New("empty message")
	}

	if payload.ChannelID == nil && payload.ReceiverID == nil {
		return events.MessageEvent{}, errors.New("no target specified")
	}

	if payload.ChannelID != nil && payload.ReceiverID != nil {
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
			ChannelID:  nil,
		}

		// return &MessageResult{
		// 	Response:   &resp,
		// 	SenderID:   senderID,
		// 	ReceiverID: payload.ReceiverID,
		// 	ChannelID:     nil,
		// }, nil
		return events, nil
	}

	// =========================
	//  CHANNEL MESSAGE
	// =========================
	channelUUID := *payload.ChannelID

	//if exist (check if exist or not )
	exists, err := s.channelRepo.ExistByID(ctx, channelUUID)
	if err != nil {
		return events.MessageEvent{}, err
	}
	if !exists {
		return events.MessageEvent{}, errors.New("receiver does not exist")
	}

	// check if a member of the channel or not ?
	isMember, err := s.channelRepo.IsMember(ctx, channelUUID, senderID)
	if err != nil {
		return events.MessageEvent{}, err
	}
	if !isMember {
		return events.MessageEvent{}, errors.New("not a member of this channel")
	}

	msg := &models.Message{
		ID:        uuid.New(),
		ChannelID: payload.ChannelID,
		SenderID:  senderID,
		Content:   &payload.Content,
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
		ChannelID:  payload.ChannelID,
	}

	return events, nil
	// resp := mapper.ToChannelMessageResponse(msg, senderID)
	// return &MessageResult{
	// 	Response:   &resp,
	// 	SenderID:   senderID,
	// 	ReceiverID: nil,
	// 	ChannelID:     payload.ChannelID,
	// }, nil
}

func (s *messageService) CreateAttachmentMessage(ctx context.Context, senderID uuid.UUID, target IncomingMessage, file AttachmentInput) (events.MessageEvent, error) {
	if target.Type != "dm" && target.Type != "channel" {
		return events.MessageEvent{}, errors.New("invalid chat type")
	}
	if target.Type == "dm" && (target.ReceiverID == nil || target.ChannelID != nil) {
		return events.MessageEvent{}, errors.New("invalid direct message target")
	}
	if target.Type == "channel" && (target.ChannelID == nil || target.ReceiverID != nil) {
		return events.MessageEvent{}, errors.New("invalid channel target")
	}
	if file.ObjectKey == "" || file.FileName == "" || file.FileSize < 0 {
		return events.MessageEvent{}, errors.New("invalid attachment metadata")
	}

	sender, err := s.userRepo.FindByID(ctx, senderID)
	if err != nil {
		return events.MessageEvent{}, err
	}
	message := &models.Message{
		ID: uuid.New(), Type: models.MessageTypeFile, SenderID: senderID,
		ReceiverID: target.ReceiverID, ChannelID: target.ChannelID,
		Content: &target.Content, ObjectKey: &file.ObjectKey,
		FileName: &file.FileName, FileSize: &file.FileSize, MimeType: &file.MimeType,
	}
	if target.ReceiverID != nil {
		if *target.ReceiverID == senderID {
			return events.MessageEvent{}, errors.New("cannot send an attachment to yourself")
		}
		exists, err := s.userRepo.ExistByID(ctx, *target.ReceiverID)
		if err != nil {
			return events.MessageEvent{}, err
		}
		if !exists {
			return events.MessageEvent{}, errors.New("recipient does not exist")
		}
	} else {
		exists, err := s.channelRepo.ExistByID(ctx, *target.ChannelID)
		if err != nil {
			return events.MessageEvent{}, err
		}
		if !exists {
			return events.MessageEvent{}, errors.New("channel does not exist")
		}
		member, err := s.channelRepo.IsMember(ctx, *target.ChannelID, senderID)
		if err != nil {
			return events.MessageEvent{}, err
		}
		if !member {
			return events.MessageEvent{}, errors.New("not a member of this channel")
		}
	}
	if err := s.messageRepo.SaveMessage(ctx, message); err != nil {
		return events.MessageEvent{}, err
	}
	return events.MessageEvent{
		ID: message.ID, Type: target.Type, Content: target.Content, CreatedAt: message.CreatedAt,
		SenderID: sender.ID, SenderName: sender.Name, ReceiverID: target.ReceiverID, ChannelID: target.ChannelID,
		FileURL:        stringPtr("/api/files/" + message.ID.String() + "/download"),
		FilePreviewURL: stringPtr("/api/files/" + message.ID.String() + "/preview"),
		FileName:       message.FileName, FileSize: message.FileSize, MimeType: message.MimeType,
	}, nil
}

func (s *messageService) GetAttachment(ctx context.Context, requesterID, messageID uuid.UUID) (*models.Message, error) {
	message, err := s.messageRepo.FindMessageByID(ctx, messageID)
	if err != nil {
		return nil, err
	}
	if message.ObjectKey == nil {
		return nil, errors.New("attachment not found")
	}
	if hiddenStore, ok := s.messageRepo.(repository.MessageDeletionRepository); ok {
		hidden, err := hiddenStore.IsMessageHiddenForUser(ctx, messageID, requesterID)
		if err != nil || hidden {
			return nil, errors.New("attachment not found")
		}
	}
	if hiddenStore, ok := s.messageRepo.(repository.MessageDeletionRepository); ok {
		hidden, err := hiddenStore.IsMessageHiddenForUser(ctx, messageID, requesterID)
		if err != nil || hidden {
			return nil, errors.New("attachment not found")
		}
	}
	if message.ReceiverID != nil {
		if requesterID != message.SenderID && requesterID != *message.ReceiverID {
			return nil, errors.New("attachment not found")
		}
		return message, nil
	}
	if message.ChannelID != nil {
		member, err := s.channelRepo.IsMember(ctx, *message.ChannelID, requesterID)
		if err != nil {
			return nil, err
		}
		if !member {
			return nil, errors.New("attachment not found")
		}
		return message, nil
	}
	return nil, errors.New("attachment not found")
}

func (s *messageService) DeleteMessage(ctx context.Context, requesterID, messageID uuid.UUID, scope string) (events.MessageDeletionEvent, *string, error) {
	message, err := s.messageRepo.FindMessageByID(ctx, messageID)
	if err != nil {
		return events.MessageDeletionEvent{}, nil, errors.New("message not found")
	}
	if scope != "me" && scope != "everyone" {
		return events.MessageDeletionEvent{}, nil, errors.New("scope must be me or everyone")
	}

	isParticipant := false
	if message.ReceiverID != nil {
		isParticipant = requesterID == message.SenderID || requesterID == *message.ReceiverID
	}
	if message.ChannelID != nil {
		isParticipant, err = s.channelRepo.IsMember(ctx, *message.ChannelID, requesterID)
		if err != nil {
			return events.MessageDeletionEvent{}, nil, err
		}
	}
	if !isParticipant {
		return events.MessageDeletionEvent{}, nil, errors.New("not authorized to delete this message")
	}

	deletionStore, ok := s.messageRepo.(repository.MessageDeletionRepository)
	if !ok {
		return events.MessageDeletionEvent{}, nil, errors.New("message deletion is unavailable")
	}
	if scope == "me" {
		if err := deletionStore.HideMessageForUser(ctx, messageID, requesterID); err != nil {
			return events.MessageDeletionEvent{}, nil, err
		}
		return events.MessageDeletionEvent{}, nil, nil
	}
	if message.SenderID != requesterID {
		return events.MessageDeletionEvent{}, nil, errors.New("only the sender can delete this message for everyone")
	}

	event := events.MessageDeletionEvent{MessageID: message.ID, SenderID: message.SenderID, ReceiverID: message.ReceiverID, ChannelID: message.ChannelID}
	objectKey := message.ObjectKey
	if err := deletionStore.DeleteMessage(ctx, messageID); err != nil {
		return events.MessageDeletionEvent{}, nil, err
	}
	return event, objectKey, nil
}

func stringPtr(value string) *string { return &value }

func (s *messageService) MarkDelivered(ctx context.Context, messageID uuid.UUID) error {
	return s.messageRepo.MarkDelivered(ctx, messageID)
}

func (s *messageService) MarkRead(ctx context.Context, userID, OtherUserID uuid.UUID) error {
	return s.messageRepo.MarkRead(ctx, userID, OtherUserID)
}
