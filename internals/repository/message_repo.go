package repository

import (
	"chat-server/internals/db/models"
	"context"
	"slices"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MessageRepository interface {
	SaveMessage(context.Context, *models.Message) error
	GetDmMessageHistory(ctx context.Context, userA string, userB string, limit int) ([]models.Message, error)
	GetCrewMessageHistory(
		ctx context.Context,
		crewID uuid.UUID,
		limit int,
		cursor *time.Time,
	) ([]models.Message, error)
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
func (r *messageRepository) GetCrewMessageHistory(
	ctx context.Context,
	crewID uuid.UUID,
	limit int,
	cursor *time.Time,
) ([]models.Message, error) {

	var messages []models.Message

	query := r.db.
		WithContext(ctx).
		Preload("Sender").
		Where("crew_id= ?", crewID)

	if cursor != nil {
		query = query.Where("created_at < ?", *cursor)
	}

	err := query.
		Order("created_at Desc").
		Limit(limit).
		Find(&messages).Error

	if err != nil {
		return nil, err
	}

	// Reverse in memory so frontend gets ASC order
	slices.Reverse(messages)

	return messages, nil
}

// saves the messages to the database
// using the ctx for leting is the user is disconnects.
// if client disconnects if WS closes if request is canceled
func (r *messageRepository) SaveMessage(ctx context.Context, msg *models.Message) error {
	return r.db.WithContext(ctx).Create(msg).Error
}

func (r *messageRepository) GetDmMessageHistory(
	ctx context.Context,
	userA string,
	userB string,
	limit int) ([]models.Message, error) {
	var messages []models.Message

	result := r.db.
		WithContext(ctx).
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
