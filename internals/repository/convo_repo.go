package repository

import (
	"chat-server/internals/db/models"
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Conversation interface {
	// Returns existing conv between two users (unordered)
	FindConvBetween(ctx context.Context, a, b uuid.UUID) (*models.Conversation, error)
	FindConvByID(ctx context.Context, id uuid.UUID) (*models.Conversation, error)
	GetMessages(ctx context.Context, convID uuid.UUID, limit, offset int) ([]models.Message, error)
	SaveMessage(ctx context.Context, msg *models.Message) error
}

type convoRepository struct {
	db *gorm.DB
}

func NewConversation(db *gorm.DB) Conversation {
	return &convoRepository{db: db}
}

// helper function to avoid duplicate tables
func orderPair(a, b uuid.UUID) (uuid.UUID, uuid.UUID) {
	if a.String() < b.String() {
		return a, b
	}
	return b, a
}

// FindconvBetween
func (r *convoRepository) FindConvBetween(ctx context.Context, a, b uuid.UUID) (*models.Conversation, error) {
	user1, user2 := orderPair(a, b)

	var convo models.Conversation

	err := r.db.WithContext(ctx).Where("user1_id = ? and user2_id = ?", user1, user2).First(&convo).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	return &convo, err
}

// FindBYID
func (r *convoRepository) FindConvByID(ctx context.Context, id uuid.UUID) (*models.Conversation, error) {
	var conv models.Conversation
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&conv).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	return &conv, err
}

// GetMessages
func (r *convoRepository) GetMessages(ctx context.Context, convID uuid.UUID, limit, offset int) ([]models.Message, error) {
	var msgs []models.Message

	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	err := r.db.
		WithContext(ctx).
		Where("conversation_id = ?", convID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&msgs).Error

	return msgs, err
}

// SaveMessage
func (r *convoRepository) SaveMessage(ctx context.Context, msg *models.Message) error {
	msg.ID = uuid.New()
	return r.db.WithContext(ctx).Create(msg).Error
}
