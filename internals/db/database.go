package db

import (
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectToDB() {
	var err error
	dbe := os.Getenv("DATABASE_URL")
	DB, err = gorm.Open(postgres.Open(dbe), &gorm.Config{
		PrepareStmt: false,
	})
	DB, err = gorm.Open(postgres.Open(dbe), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to Connect to the DATABASE")
	}
	log.Println("DATABASE Connection sucessful.")
	DB.AutoMigrate(&User{}, &Account{}, &Session{}, &Product{}, &Cart{}, &CartItem{})
	log.Printf("Database migrated sucessfully")
}

// FindUserByID retrieves a user from the database by their ID.
func FindUserByID(userID string) (*User, error) {
	var user User
	// Use GORM's .First() method to find the record.
	// "id = ?" is a secure way to query, preventing SQL injection.
	result := DB.First(&user, "id = ?", userID)

	if result.Error != nil {
		// Check if the error is because the record was not found.
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil // Return nil, nil to indicate "not found" without an error.
		}
		// For any other database error, return the error.
		return nil, result.Error
	}

	return &user, nil
}
