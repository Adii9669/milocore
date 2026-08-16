package models

import (
	"time"

	"github.com/google/uuid"
)

type MessageType string

const (
	MessageTypeText  MessageType = "text"
	MessageTypeImage MessageType = "image"
	MessageTypeAudio MessageType = "audio"
	MessageTypeVideo MessageType = "video"
	MessageTypeFile  MessageType = "file"
)

type Message struct {
	ID   uuid.UUID   `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Type MessageType `gorm:"type:varchar(20);not null;default:'text';index"`

	Sender     User       `gorm:"foreignKey:SenderID;references:ID"`
	ReceiverID *uuid.UUID `gorm:"type:uuid;index:idx_receiver_created,priority:1"`
	SenderID   uuid.UUID  `gorm:"type:uuid;not null;idx_sender_created"`
	CrewID     *uuid.UUID `gorm:"type:uuid;index:idx_crew_created,priority:1"`

	Content   *string `gorm:"type:text"`
	Delivered bool    `gorm:"default:false;index"`
	Read      bool    `gorm:"default:false;index"`

	CreatedAt time.Time `gorm:"index:idx_crew_created,priority:2;index:idx_receiver_created,priority:2"`
	UpdatedAt time.Time

	FileURL  *string
	FileName *string
	FileSize *int64
	MimeType *string
}
