package models

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	SenderName string    `gorm:"type:text"`

	SenderID   string  `gorm:"type:uuid;not null;idx_sender_created"`
	ReceiverID *string `gorm:"type:uuid;idx_receiver_created"`

	CrewID *string `gorm:"type:uuid;index:idx_crew_chat,priority:1"`

	Content   *string `gorm:"type:text;not null"`
	Delivered bool    `gorm:"default:false;index"`
	Read      bool    `gorm:"default:false;index"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
