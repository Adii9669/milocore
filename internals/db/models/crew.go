package models

import (
	"github.com/google/uuid"
	"time"
)

// crew
type Crew struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Name        string    `gorm:"size:100;not null"`
	Description string    `gorm:"type:text"`
	IconURL     string    `gorm:"type:varchar(255)"`

	//The person who is created the crew
	OwnerID uuid.UUID `gorm:"type:uuid;not null"`

	//Metadata
	CreatedAt time.Time
	UpdateAt  time.Time

	//Relation (for GORM to fetch data)
	Owner    User      `gorm:"foreignKey:OwnerID"`
	Members  []User    `gorm:"many2many:crew_members;constraint:OnDelete:CASCADE;"`
	Messages []Message `gorm:"foreignKey:CrewID;constraint:OnDelete:CASCADE;"`
}
