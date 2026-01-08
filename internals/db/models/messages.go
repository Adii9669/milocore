package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type Message struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	SenderID    string    `gorm:"type:uuid;not null;idx_sender_created"`
	ReceiverID  *string   `gorm:"type:uuid;idx_receiver_created"`
	CrewID      *string   `gorm:"type:uuid;index;idx_crew_created"`
	Content     *string   `gorm:"type:text;"`
	ContentType string    `gorm:"size:50;not null;default;'text'"`
	FileURL     *string   `gorm:"type:text;"`
	FileName    *string   `gorm:"size:255;"`
	FileSize    *int64
	Metadata    datatypes.JSONMap `gorm:"type:jsonb"`
	Delivered   bool              `gorm:"default:false;index"`
	Read        bool              `gorm:"default:false;index"`
	CreatedAt   time.Time
	UpdateAt    time.Time
}
