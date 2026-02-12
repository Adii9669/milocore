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

func ConnectToDB() *gorm.DB {

	dbe := os.Getenv("DATABASE_URL")
	// runMigrations := os.Getenv("RUN_MIGRATIONS")
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
	runMig := false

	if runMig {
		if os.Getenv("APP_ENV") == "local" {
			if err := DB.AutoMigrate(
				&models.User{},
				&models.Account{},
				&models.Session{},
				&models.Crew{},
				&models.Message{},
				&models.Friends{},
			); err != nil {
				log.Fatalf("Migration failed: %v", err)
			}
			log.Println("Database migration successful")
		} else {
			log.Println("-----Skiping the migration-------")
		}

	}

	return DB
}

// FindUserByID retrieves a user from the database by their ID.
