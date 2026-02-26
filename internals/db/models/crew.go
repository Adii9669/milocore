package models

import (
	"github.com/google/uuid"
	"time"
)

type CrewRole string

const (
	RoleOwner  CrewRole = "owner"
	RoleAdmin  CrewRole = "admin"
	RoleMember CrewRole = "member"
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
	UpdatedAt time.Time

	//Relation (for GORM to fetch data)
	Owner    User         `gorm:"foreignKey:OwnerID"`
	Members  []CrewMember `gorm:"foreignKey:CrewID;constraint:OnDelete:CASCADE;"`
	Messages []Message    `gorm:"foreignKey:CrewID;constraint:OnDelete:CASCADE;"`
}

type CrewMember struct {
	CrewID uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID uuid.UUID `gorm:"type:uuid;primaryKey"`

	Role     CrewRole  `gorm:"type:varchar(20);not null;default:member"`
	JoinedAt time.Time `gorm:"autoCreateTime"`
	MutedTil *time.Time

	Crew Crew `gorm:"foreignKey:CrewID"`
	User User `gorm:"foreignKey:UserID"`
}

func (r CrewRole) IsValid() bool {
	switch r {
	case RoleOwner, RoleAdmin, RoleMember:
		return true
	default:
		return false
	}
}
