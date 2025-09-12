package db

import (
	"time"
)

type Product struct {
	ID          string `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Description string
	ImageURL    string
	Name        string
	Price       float64 // FIX: Price should be a numeric type for calculations.
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CartItems   []CartItem `gorm:"foreignKey:ProductID"` // FIX: Field name is now plural for clarity.
}

type Cart struct {
	ID        string  `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID    *string `gorm:"type:uuid;unique"`
	CreatedAt time.Time
	UpdatedAt time.Time
	User      *User      `gorm:"foreignKey:UserID"`
	Items     []CartItem `gorm:"foreignKey:CartID"`
}

type CartItem struct {
	ID        string `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ProductID string `gorm:"type:uuid"`
	Quantity  int
	CartID    string  `gorm:"type:uuid"`
	Cart      Cart    `gorm:"foreignKey:CartID"`
	Product   Product `gorm:"foreignKey:ProductID"`
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
	User              User `gorm:"foreignKey:UserID"`
}

type Session struct {
	ID           uint   `gorm:"primaryKey"`
	SessionToken string `gorm:"unique"`
	UserID       string `gorm:"type:uuid"`
	Expires      time.Time
	User         User `gorm:"foreignKey:UserID"`
}

type User struct {
	ID           string  `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name         *string `gorm:"unique"`
	PasswordHash *string
	Email        *string `gorm:"uniqueIndex"`
	Image        *string
	Verified     bool    `gorm:"default:false"`
	VerifyToken  *string `gorm:"column:verify_token"`
	UpdatedAt    time.Time
	CreatedAt    time.Time
	Cart         *Cart     `gorm:"foreignKey:UserID"`
	Accounts     []Account `gorm:"foreignKey:UserID"`
	Sessions     []Session `gorm:"foreignKey:UserID"`
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
