package utils

import (
	"log"
	"time"

	"chat-server/internals/db"
	"chat-server/internals/db/models"
)

func CleanupUnverifiedUsers() {
	// users older than 7 days AND not verified
	cutoff := time.Now().Add(-7 * 24 * time.Hour) // 7 days ago

	result := db.DB.
		Where("verified = ? AND created_at < ?", false, cutoff).
		Delete(&models.User{})

	if result.Error != nil {
		log.Printf("Error cleaning up unverified users: %v", result.Error)
		return
	}

	log.Printf("Deleted %d unverified users older than 7 days.", result.RowsAffected)
}
