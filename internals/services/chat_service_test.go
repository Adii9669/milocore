package services

import (
	"chat-server/internals/db/models"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type MockMessageRepo struct {
	GetCrewMessageHistoryFn func(ctx context.Context, crewID uuid.UUID, limit int, cursor *time.Time) ([]models.Message, error)
	GetDmMessageHistoryFn   func(ctx context.Context, userA, userB uuid.UUID, limit int, cursor *time.Time) ([]models.Message, error)
	MarkDeliveredFn         func(ctx context.Context, messageID uuid.UUID) error
	MarkReadFn              func(ctx context.Context, userID, otherUserID uuid.UUID) error
	SaveMessageFn           func(context.Context, *models.Message) error
}

func (m *MockMessageRepo) SaveMessage(ctx context.Context, msg *models.Message) error {
	if m.SaveMessageFn != nil {
		return m.SaveMessageFn(ctx, msg)
	}
	return nil
}

func (m *MockMessageRepo) MarkRead(ctx context.Context, userID, otherUserID uuid.UUID) error {
	if m.MarkReadFn != nil {
		return m.MarkReadFn(ctx, userID, otherUserID)
	}
	return nil
}

func (m *MockMessageRepo) MarkDelivered(ctx context.Context, messageID uuid.UUID) error {
	if m.MarkDeliveredFn != nil {
		return m.MarkDeliveredFn(ctx, messageID)
	}
	return nil
}

func (m *MockMessageRepo) GetCrewMessageHistory(
	ctx context.Context,
	crewID uuid.UUID,
	limit int,
	cursor *time.Time,
) ([]models.Message, error) {

	if m.GetCrewMessageHistoryFn != nil {
		return m.GetCrewMessageHistoryFn(ctx, crewID, limit, cursor)
	}
	return nil, nil
}
func (m *MockMessageRepo) GetDmMessageHistory(ctx context.Context, userA, userB uuid.UUID, limit int, cursor *time.Time) ([]models.Message, error) {
	if m.GetDmMessageHistoryFn != nil {
		return m.GetDmMessageHistoryFn(ctx, userA, userB, limit, cursor)
	}
	return nil, nil
}

// Testing the Success
func TestCrewHistory_Success(t *testing.T) {
	ctx := context.Background()
	crewID := uuid.New()
	userID := uuid.New()

	mockRepo := &MockMessageRepo{
		GetCrewMessageHistoryFn: func(ctx context.Context, crewID uuid.UUID, limit int, cursor *time.Time) ([]models.Message, error) {

			// 🔥 important: verify limit logic
			assert.Equal(t, 50, limit)

			message := "hello"
			return []models.Message{
				{
					ID:      uuid.New(),
					Content: &message,
				},
			}, nil
		},
	}

	service := NewChatHistoryService(mockRepo)

	result, err := service.CrewHistory(ctx, crewID, userID, 0, nil)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestCrewHistory_Faliure(t *testing.T) {
	ctx := context.Background()

	mockRepo := &MockMessageRepo{
		GetCrewMessageHistoryFn: func(ctx context.Context, crewID uuid.UUID, limit int, cursor *time.Time) ([]models.Message, error) {
			return nil, errors.New("db error")
		},
	}
	service := NewChatHistoryService(mockRepo)

	result, err := service.CrewHistory(ctx, uuid.New(), uuid.New(), 10, nil)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestDMHistory_InvalidUser(t *testing.T) {
	services := NewChatHistoryService(nil)

	_, err := services.DmHistory(context.Background(), uuid.Nil, uuid.New(), 10, nil)

	assert.Error(t, err)

}

func TestDMSame_User(t *testing.T) {
	id := uuid.New()

	services := NewChatHistoryService(nil)

	_, err := services.DmHistory(context.Background(), id, id, 10, nil)

	assert.Error(t, err)
}

func TestDmHistory_Success(t *testing.T) {
	ctx := context.Background()
	userA := uuid.New()
	userB := uuid.New()

	mockRepo := &MockMessageRepo{
		GetDmMessageHistoryFn: func(ctx context.Context, a, b uuid.UUID, limit int, cursor *time.Time) ([]models.Message, error) {

			assert.Equal(t, 50, limit) // limit correction

			message := "hello"
			return []models.Message{
				{
					ID:      uuid.New(),
					Content: &message,
				},
			}, nil
		},
	}

	service := NewChatHistoryService(mockRepo)

	result, err := service.DmHistory(ctx, userA, userB, 0, nil)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestDMHistory_Faliure(t *testing.T) {
	ctx := context.Background()
	userA := uuid.New()
	userB := uuid.New()

	mockRepo := &MockMessageRepo{
		GetDmMessageHistoryFn: func(ctx context.Context, a, b uuid.UUID, limit int, cursor *time.Time) ([]models.Message, error) {
			return nil, errors.New("db error")
		},
	}

	service := NewChatHistoryService(mockRepo)

	result, err := service.DmHistory(ctx, userA, userB, 0, nil)

	assert.Error(t, err)  // ✅ expect error
	assert.Nil(t, result) // ✅ no result

}
