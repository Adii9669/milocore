package models

import (
	"github.com/google/uuid"
	"time"
)

// crew
type Crew struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Name      string    `gorm:"size:100;not null"`
	OwnerID   uuid.UUID `gorm:"type:uuid;not null"`
	CreatedAt time.Time

	Owner    User      `gorm:"foreignKey:OwnerID"`
	Members  []User    `gorm:"many2many:crew_members;constraint:OnDelete:CASCADE;"`
	Messages []Message `gorm:"foreignKey:CrewID;constraint:OnDelete:CASCADE;"`
}
