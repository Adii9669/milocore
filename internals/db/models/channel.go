package models

import (
	"time"

	"github.com/google/uuid"
)

type ChannelRole string

const (
	RoleOwner  ChannelRole = "owner"
	RoleAdmin  ChannelRole = "admin"
	RoleMember ChannelRole = "member"
)

// channel
type Channel struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Name        string    `gorm:"size:100;not null"`
	Description string    `gorm:"type:text"`
	IconURL     string    `gorm:"type:varchar(255)"`

	//The person who is created the channel
	OwnerID uuid.UUID `gorm:"type:uuid;not null"`

	//Metadata
	CreatedAt time.Time
	UpdatedAt time.Time

	//Relation (for GORM to fetch data)
	Owner    User            `gorm:"foreignKey:OwnerID"`
	Members  []ChannelMember `gorm:"foreignKey:ChannelID;constraint:OnDelete:CASCADE;"`
	Messages []Message       `gorm:"foreignKey:ChannelID;constraint:OnDelete:CASCADE;"`
}

type ChannelMember struct {
	ChannelID uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey"`

	Role     ChannelRole `gorm:"type:varchar(20);not null;default:member"`
	JoinedAt time.Time   `gorm:"autoCreateTime"`
	MutedTil *time.Time

	Channel Channel `gorm:"foreignKey:ChannelID"`
	User    User    `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

func (r ChannelRole) IsValid() bool {
	switch r {
	case RoleOwner, RoleAdmin, RoleMember:
		return true
	default:
		return false
	}
}
