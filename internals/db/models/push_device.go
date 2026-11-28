package models

import (
	"time"

	"github.com/google/uuid"
)

type PushDevice struct {
	InstallationID uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID         uuid.UUID `gorm:"type:uuid;not null;index"`
	User           User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Token          string    `gorm:"type:text;not null;uniqueIndex"`
	UpdatedAt      time.Time
}
