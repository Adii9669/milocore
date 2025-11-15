package models

import (
	"time"
)

type User struct {
	ID               string  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Name             *string `gorm:"unique"`
	PasswordHash     *string
	Email            *string `gorm:"uniqueIndex"`
	Image            *string
	Verified         bool    `gorm:"default:false"`
	VerifyOTP        *string `gorm:"column:verify_otp"`
	VerificationCode string
	UpdatedAt        time.Time
	CreatedAt        time.Time

	//relation
	Accounts   []Account `gorm:"foreignKey:UserID"`
	Sessions   []Session `gorm:"foreignKey:UserID"`
	OwnedCrews []Crew    `gorm:"foreignKey:OwnerID"`
	Crews      []Crew    `gorm:"many2many:crew_members;constraint:OnDelete:CASCADE;"`
	Messages   []Message `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"`
}

type Account struct {
	ID                string `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID            string `gorm:"type:uuid"`
	Type              string
	Provider          string `gorm:"uniqueIndex:idx_provider_account"`
	ProviderAccountId string `gorm:"uniqueIndex:idx_provider_account"`
	RefreshToken      *string
	AccessToken       *string
	ExpiresAt         *int
	TokenType         *string
	Scope             *string
	IDToken           *string `gorm:"column:id_token"`
	SessionState      *string

	//relation
	User User `gorm:"foreignKey:UserID"`
}

type Session struct {
	ID           uint   `gorm:"primaryKey"`
	SessionToken string `gorm:"unique"`
	UserID       string `gorm:"type:uuid"`
	Expires      time.Time

	//relation
	User User `gorm:"foreignKey:UserID"`
}

// GetID should implement the interface by returning the user's ID.
// An ID is almost never optional, so we return a plain string.
func (u User) GetID() string {
	return u.ID
}

// GetName should implement the interface by returning the user's name.
func (u User) GetName() *string {
	return u.Name
}
