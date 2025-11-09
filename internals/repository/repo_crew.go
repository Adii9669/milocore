package repository

import (
	"chat-server/internals/db"
	"chat-server/internals/db/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CrewRepository interface {
	CreateCrew(crew *models.Crew) error
	FindForUser(userID string) ([]models.Crew, error)
	// FindUserID(userID uuid.UUID) ([]models.Crew, error)
	DeleteCrewByID(ownerID uuid.UUID, crewID uuid.UUID) error
}

type GormCrewRepository struct{}

// create cerw
func (r *GormCrewRepository) CreateCrew(crew *models.Crew) error {

	return db.DB.Transaction(func(tx *gorm.DB) error {
		//1. Create a Crew
		if err := tx.Create(crew).Error; err != nil {
			return err
		}

		//2.Fetch the owner
		var owner models.User
		if err := tx.First(&owner, "id=?", crew.OwnerID).Error; err != nil {
			return err
		}

		// 3. Add owner as member using GORM association (auto-handles join table)
		if err := tx.Model(crew).Association("Members").Append(&owner); err != nil {
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

// Delete the crews
func (r *GormCrewRepository) DeleteCrewByID(ownerID uuid.UUID, crewID uuid.UUID) error {
	return db.DB.Transaction(func(tx *gorm.DB) error {

		//1. chec owner ship
		var crew models.Crew
		if err := tx.Where("id=? AND owner_id = ?", crewID, ownerID).First(&crew).Error; err != nil {
			return err
		}

		//2.Delete Crew
		// (A) Manual cleanup if you DON'T have ON DELETE CASCADE:
		// if err := tx.Exec("DELETE FROM crew_members WHERE crew_id = ?", crewID).Error; err != nil {
		// 	return err
		// }
		// if err := tx.Exec("DELETE FROM messages WHERE crew_id = ?", crewID).Error; err != nil {
		// 	return err
		// }

		//3.Delete Crew
		if err := tx.Delete(&crew).Error; err != nil {
			return err

		}
		return nil
	})
}
