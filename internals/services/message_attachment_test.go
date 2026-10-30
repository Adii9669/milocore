package services

import (
	"chat-server/internals/db/models"
	"chat-server/internals/repository"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type attachmentUserRepoStub struct {
	repository.UserRepository
	user *models.User
}

func (s attachmentUserRepoStub) FindByID(context.Context, uuid.UUID) (*models.User, error) {
	return s.user, nil
}
func (s attachmentUserRepoStub) ExistByID(context.Context, uuid.UUID) (bool, error) {
	return true, nil
}

type attachmentChannelRepoStub struct {
	repository.ChannelRepository
	member bool
}

func (s attachmentChannelRepoStub) ExistByID(context.Context, uuid.UUID) (bool, error) {
	return true, nil
}
func (s attachmentChannelRepoStub) IsMember(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
	return s.member, nil
}

func TestCreateAttachmentMessagePersistsObjectMetadata(t *testing.T) {
	senderID, recipientID := uuid.New(), uuid.New()
	saved := (*models.Message)(nil)
	repo := &MockMessageRepo{SaveMessageFn: func(_ context.Context, message *models.Message) error {
		saved = message
		return nil
	}}
	user := &models.User{ID: senderID, Name: "sender"}
	service := NewMessageService(repo, attachmentUserRepoStub{user: user}, attachmentChannelRepoStub{})

	event, err := service.CreateAttachmentMessage(context.Background(), senderID, IncomingMessage{
		Type: "dm", ReceiverID: &recipientID, Content: "caption",
	}, AttachmentInput{ObjectKey: "users/owner/random", FileName: "notes.pdf", FileSize: 42, MimeType: "application/pdf"})

	require.NoError(t, err)
	require.NotNil(t, saved)
	require.Equal(t, models.MessageTypeFile, saved.Type)
	require.Equal(t, "users/owner/random", *saved.ObjectKey)
	require.Equal(t, "notes.pdf", *saved.FileName)
	require.EqualValues(t, 42, *saved.FileSize)
	require.Equal(t, "caption", *saved.Content)
	require.Equal(t, "dm", event.Type)
	require.Equal(t, recipientID, *event.ReceiverID)
	require.Equal(t, "/api/files/"+event.ID.String()+"/download", *event.FileURL)
}

func TestGetAttachmentRequiresChatMembership(t *testing.T) {
	senderID, recipientID, outsiderID, messageID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	objectKey := "users/sender/random"
	repo := &MockMessageRepo{FindMessageByIDFn: func(context.Context, uuid.UUID) (*models.Message, error) {
		return &models.Message{ID: messageID, SenderID: senderID, ReceiverID: &recipientID, ObjectKey: &objectKey}, nil
	}}
	service := NewMessageService(repo, attachmentUserRepoStub{}, attachmentChannelRepoStub{})

	message, err := service.GetAttachment(context.Background(), recipientID, messageID)
	require.NoError(t, err)
	require.Equal(t, messageID, message.ID)

	_, err = service.GetAttachment(context.Background(), outsiderID, messageID)
	require.Error(t, err)
}

func TestCreateChannelAttachmentRequiresMembership(t *testing.T) {
	senderID, channelID := uuid.New(), uuid.New()
	repo := &MockMessageRepo{}
	service := NewMessageService(repo, attachmentUserRepoStub{user: &models.User{ID: senderID}}, attachmentChannelRepoStub{member: false})

	_, err := service.CreateAttachmentMessage(context.Background(), senderID, IncomingMessage{
		Type: "channel", ChannelID: &channelID,
	}, AttachmentInput{ObjectKey: "users/sender/random", FileName: "file.bin"})

	require.Error(t, err)
}
