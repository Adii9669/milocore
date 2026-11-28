package models

import (
	"github.com/google/uuid"
	"time"
)

type ChannelInvite struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	ChannelID   uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_pending_channel_invite,where:status = 'pending'"`
	InviterID   uuid.UUID `gorm:"type:uuid;not null"`
	InviteeID   uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_pending_channel_invite,where:status = 'pending'"`
	Status      string    `gorm:"not null"`
	CreatedAt   time.Time
	ExpiresAt   time.Time `gorm:"not null"`
	RespondedAt *time.Time
	Channel     Channel `gorm:"constraint:OnDelete:CASCADE"`
}
