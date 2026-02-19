package models

import (
	"time"

	"github.com/google/uuid"
)

type Friend struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	UserLowID  uuid.UUID `gorm:"uniqueIndex:idx_pair;index"`
	UserHighID uuid.UUID `gorm:"uniqueIndex:idx_pair;index"`

	UserLow  User `gorm:"foreignKey:UserLowID;constraint:OnDelete:CASCADE"`
	UserHigh User `gorm:"foreignKey:UserHighID;constraint:OnDelete:CASCADE"`

	Status      string    `gorm:"type:varchar(20);not null;index"`
	RequestedBy uuid.UUID `gorm:"type:uuid;not null"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

// type Friends struct {
// 	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
//
// 	UserID   uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_user_friend"`
// 	FriendID uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_user_friend"`
// 	Status   string    `gorm:"type:varchar(20);default:'pending'"`
//
// 	User   User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
// 	Friend User `gorm:"foreignKey:FriendID;constraint:OnDelete:CASCADE"`
//
// 	CreatedAt time.Time
// 	UpdatedAt time.Time
// }
