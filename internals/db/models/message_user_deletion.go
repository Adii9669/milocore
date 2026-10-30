package models

import (
	"time"

	"github.com/google/uuid"
)

// MessageUserDeletion records a message hidden from one participant only.
type MessageUserDeletion struct {
	MessageID uuid.UUID `gorm:"type:uuid;primaryKey;index"`
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey;index"`
	CreatedAt time.Time
}
