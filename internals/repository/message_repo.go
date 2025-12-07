package repository

import (
	"chat-server/internals/db"
	"chat-server/internals/db/models"
)

type MessageRepository interface {
	GetMessagesByCrewID(crewID string, limit int) ([]models.Message, error)
	SaveMessage(msg *models.Message) error
	GetMessagesByUserID(userID string, limit int) ([]models.Message, error)
}

type messageRepository struct{}

// constructor
func NewMessageRepository() MessageRepository {
	return &messageRepository{}
}

func (r *messageRepository) GetMessagesByCrewID(crewID string, limit int) ([]models.Message, error) {
	var messages []models.Message

	result := db.DB.Where("crew_id= ?", crewID).Order("created_at asc").Limit(limit).Find(&messages)

	return messages, result.Error

}

func (r *messageRepository) SaveMessage(msg *models.Message) error {
	return db.DB.Create(msg).Error
}

func (r *messageRepository) GetMessagesByUserID(userID string, limit int) ([]models.Message, error) {
	var messages []models.Message

	result := db.DB.Where("user_id = ? ", userID).Order("created_at asc").Limit(limit).Find(&messages)

	return messages, result.Error
}
