package db

import (
	"chat-server/internals/db/models"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
var DB *gorm.DB

func ConnectToDB() {
	dbe := os.Getenv("DATABASE_URL")
	runMigrations := os.Getenv("RUN_MIGRATIONS")
	if dbe == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	// -- ConnectToDB Database ------
	var err error
	DB, err = gorm.Open(postgres.Open(dbe), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to Connect to the DATABASE")
	}
	log.Println("DATABASE Connection sucessful.")

	//---Run AutoMigrate ---
	if runMigrations == "true" {

		if err := DB.AutoMigrate(
			&models.User{},
			&models.Account{},
			&models.Session{},
			&models.Product{},
			&models.Crew{},
			&models.Message{},
		); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		log.Println("Database migration successful")
	} else {

		log.Println("-----Skiping the migration-------")
	}
}

// FindUserByID retrieves a user from the database by their ID.
func FindUserByID(userID string) (*models.User, error) {
	var user models.User
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
