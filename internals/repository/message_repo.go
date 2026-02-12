package repository

import (
	"chat-server/internals/db/models"
	"context"

	"gorm.io/gorm"
)

type MessageRepository interface {
	SaveMessage(context.Context, *models.Message) error
	GetDmMessageHistory(userA string, userB string, limit int) ([]models.Message, error)
	GetCrewMessageHistory(crewID string, limit int) ([]models.Message, error)
}

type messageRepository struct {
	db *gorm.DB
}

// constructor
func NewMessageRepository(db *gorm.DB) MessageRepository {
	return &messageRepository{db: db}
}

// What this does
// Fetches group chat history
// Sorted oldest → newest
// Limit applied
func (r *messageRepository) GetCrewMessageHistory(crewID string, limit int) ([]models.Message, error) {
	var messages []models.Message
	result := r.db.
		Preload("Sender").
		Where("crew_id= ?", crewID).
		Order("created_at asc").
		Limit(limit).
		Find(&messages)

	return messages, result.Error
}

// saves the messages to the database
// using the ctx for leting is the user is disconnects.
// if client disconnects if WS closes if request is canceled
func (r *messageRepository) SaveMessage(ctx context.Context, msg *models.Message) error {
	return r.db.WithContext(ctx).Create(msg).Error
}

func (r *messageRepository) GetDmMessageHistory(
	userA string,
	userB string,
	limit int) ([]models.Message, error) {
	var messages []models.Message

	result := r.db.
		Preload("Sender").
		Where(
			"(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
			userA, userB,
			userB, userA,
		).
		Order("created_at ASC").
		Limit(limit).
		Find(&messages)

	return messages, result.Error
}
