package store

import (
	"chat-server/internals/db/models"
	"context"
	"time"

	"gorm.io/gorm"
)

type Store struct {
	DB *gorm.DB
}

func New(db *gorm.DB) *Store {
	return &Store{
		DB: db,
	}
}

// --------------------{Helper Function}----------------------
func (s *Store) SaveMessage(ctx context.Context, m *models.Message) (*models.Message, error) {
	m.CreatedAt = time.Now().UTC()
	if err := s.DB.WithContext(ctx).Create(m).Error; err != nil {
		return nil, err
	}
	return m, nil
}

// MarkMessageRead sets read=true for a message by id.
func (s *Store) MarkMessageRead(ctx context.Context, messageID string) error {
	return s.DB.WithContext(ctx).Model(&models.Message{}).Where("id = ?", messageID).Update("read", true).Error
}

// IsUserCrewMember checks if user is a member of the crew.
func (s *Store) IsUserCrewMember(ctx context.Context, userID, crewID string) (bool, error) {
	var c int64
	if err := s.DB.WithContext(ctx).
		Model(&models.Crew{}).
		Where("user_id = ? AND crew_id = ?", userID, crewID).
		Count(&c).Error; err != nil {
		return false, err
	}
	return c > 0, nil
}
