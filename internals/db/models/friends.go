package models

import (
	"time"

	"github.com/google/uuid"
)

type Friends struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	UserID    uuid.UUID `gorm:"type:uuid;index"`
	FriendID  uuid.UUID `gorm:"type:uuid;index"`
	Status    string    `gorm:"type:varchar(20);default:'pending'"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
