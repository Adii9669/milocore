package models

import (
	"github.com/google/uuid"
	"time"
)

type Message struct {
	ID        uuid.UUID `gorm:"type:uuid;primarKey;default:uuid_generate_v4()"`
	Content   string    `gorm:"type:text;not null"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;"`
	CrewID    uuid.UUID `gorm:"type:uuid;not null;"`
	CreatedAt time.Time

	//relations
	User User `gorm:"foreignKey:UserID"`
	Crew Crew `gorm:"foreignKey:UserID"`
}
