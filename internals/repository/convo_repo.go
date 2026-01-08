package repository

import (
	"chat-server/internals/db"
	"chat-server/internals/db/models"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Conversation interface {
	// Returns existing conv between two users (unordered)
	FindConvBetween(a, b uuid.UUID) (*models.Conversation, error)
	FindConvByID(id uuid.UUID) (*models.Conversation, error)
	GetMessages(convID uuid.UUID, limit, offset int) ([]models.Message, error)
	SaveMessage(msg *models.Message) error
}

type convoRepository struct{}

func NewConversation() Conversation {
	return &convoRepository{}
}

// helper function to avoid duplicate tables
func orderPair(a, b uuid.UUID) (uuid.UUID, uuid.UUID) {
	if a.String() < b.String() {
		return a, b
	}
	return b, a
}

// FindconvBetween
func (r *convoRepository) FindConvBetween(a, b uuid.UUID) (*models.Conversation, error) {
	user1, user2 := orderPair(a, b)

	var convo models.Conversation

	err := db.DB.Where("user1_id = ? and user2_id = ?", user1, user2).First(&convo).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	return &convo, err
}

// FindBYID
func (r *convoRepository) FindConvByID(id uuid.UUID) (*models.Conversation, error) {
	var conv models.Conversation
	err := db.DB.Where("id = ?", id).First(&conv).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	return &conv, err
}

// GetMessages
func (r *convoRepository) GetMessages(convID uuid.UUID, limit, offset int) ([]models.Message, error) {
	var msgs []models.Message

	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	err := db.DB.
		Where("conversation_id = ?", convID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&msgs).Error

	return msgs, err
}

// SaveMessage
func (r *convoRepository) SaveMessage(msg *models.Message) error {
	msg.ID = uuid.New()
	return db.DB.Create(msg).Error
}
