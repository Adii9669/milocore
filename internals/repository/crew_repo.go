package repository

import (
	"chat-server/internals/db"
	"chat-server/internals/db/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CrewRepository interface {
	Create(crew *models.Crew) error
	FindForUser(userID string) ([]models.Crew, error)
	FindUserID(userID uuid.UUID) ([]models.Crew, error)
}

type GormCrewRepository struct{}

// create cerw
func (r *GormCrewRepository) Create(crew *models.Crew) error {

	return db.DB.Transaction(func(tx *gorm.DB) error {
		//1. Create a Crew
		if err := tx.Create(crew).Error; err != nil {
			return err
		}

		//2.
		if err := tx.Exec("INSERT INTO crew_members (crew_id, user_id) VALUES (?, ?)", crew.ID, crew.OwnerID).Error; err != nil {
			return err
		}
		return nil
	})

}

// for FindForUser
func (r *GormCrewRepository) FindForUser(userID string) ([]models.Crew, error) {
	var crews []models.Crew
	var user models.User

	// 1. First, find the user to start the association from.
	if err := db.DB.First(&user, "id = ?", userID).Error; err != nil {
		return nil, err
	}

	// 2. Now, use Association to find all the 'Crews' linked to that user.
	err := db.DB.Model(&user).Association("Crews").Find(&crews)
	return crews, err
}
