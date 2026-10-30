package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	PlanStatusOpen      = "open"
	PlanStatusConfirmed = "confirmed"
	PlanStatusCompleted = "completed"
	PlanStatusCancelled = "cancelled"
)

func IsValidPlanStatus(status string) bool {
	switch status {
	case PlanStatusOpen, PlanStatusConfirmed, PlanStatusCompleted, PlanStatusCancelled:
		return true
	default:
		return false
	}
}

type Plan struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Title        string     `gorm:"size:200;not null;check:chk_plans_title_nonblank,length(btrim(title)) > 0" json:"title"`
	Description  *string    `gorm:"type:text" json:"description,omitempty"`
	CreatorID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"creatorId"`
	ChannelID    *uuid.UUID `gorm:"type:uuid;index;check:chk_plans_chat_target,(channel_id IS NOT NULL AND dm_user_low_id IS NULL AND dm_user_high_id IS NULL) OR (channel_id IS NULL AND dm_user_low_id IS NOT NULL AND dm_user_high_id IS NOT NULL)" json:"channelId,omitempty"`
	DMUserLowID  *uuid.UUID `gorm:"type:uuid;index" json:"dmUserLowId,omitempty"`
	DMUserHighID *uuid.UUID `gorm:"type:uuid;index" json:"dmUserHighId,omitempty"`
	StartAt      time.Time  `gorm:"type:timestamptz;not null;index" json:"startAt"`
	EndAt        *time.Time `gorm:"type:timestamptz" json:"endAt,omitempty"`
	Location     *string    `gorm:"size:255" json:"location,omitempty"`
	Status       string     `gorm:"size:24;not null;default:open;index;check:chk_plans_status,status = 'open' OR status = 'confirmed' OR status = 'completed' OR status = 'cancelled'" json:"status"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`

	Creator    User     `gorm:"foreignKey:CreatorID;constraint:OnDelete:CASCADE" json:"-"`
	Channel    *Channel `gorm:"foreignKey:ChannelID;constraint:OnDelete:CASCADE" json:"-"`
	DMUserLow  *User    `gorm:"foreignKey:DMUserLowID;constraint:OnDelete:CASCADE" json:"-"`
	DMUserHigh *User    `gorm:"foreignKey:DMUserHighID;constraint:OnDelete:CASCADE" json:"-"`
}
