package models

import (
	"time"

	"gorm.io/gorm"
)

type Conversation struct {
	ID string `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	// "dm" | "group" | "crew_link"(optional) optional display name (for group DMs); for 1:1 can be null

	Type          string    `gorm:"size:30;not null;default:'dm'" json:"type"`
	Title         *string   `gorm:"size:255" json:"title,omitempty"`
	LastMessageID *string   `gorm:"type:uuid;index" json:"last_message_id,omitempty"`
	LastUpdatedAt time.Time `gorm:"index" json:"last_updated_at"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

type ConversationMember struct {
	ID             string `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	ConversationID string `gorm:"type:uuid;not null;index:idx_conv_member_conv" json:"conversation_id"`
	UserID         string `gorm:"type:uuid;not null;index:idx_conv_member_user" json:"user_id"`
	// unread counter for that user in that conversation
	UnreadCount       int     `gorm:"default:0;not null" json:"unread_count"`
	LastReadMessageID *string `gorm:"type:uuid;index" json:"last_read_message_id,omitempty"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}
