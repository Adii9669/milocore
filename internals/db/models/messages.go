package models

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`

	SenderID   uuid.UUID  `gorm:"type:uuid;not null;idx_sender_created"`
	Sender     User       `gorm:"foreignKey:SenderID;references:ID"`
	ReceiverID *uuid.UUID `gorm:"type:uuid;index:idx_receiver_created,priority:1"`

	CrewID *uuid.UUID `gorm:"type:uuid;index:idx_crew_created,priority:1"`

	Content   *string `gorm:"type:text;not null"`
	Delivered bool    `gorm:"default:false;index"`
	Read      bool    `gorm:"default:false;index"`

	CreatedAt time.Time `gorm:"index:idx_crew_created,priority:2;index:idx_receiver_created,priority:2"`
	UpdatedAt time.Time
}
