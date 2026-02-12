package repository

import (
	"chat-server/internals/db/models"
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CrewRepository interface {
	CreateCrew(crew *models.Crew) error
	FindForUser(userID string) ([]models.Crew, error)
	DeleteCrewByID(ownerID uuid.UUID, crewID uuid.UUID) error
	ExistByID(ctx context.Context, crewID uuid.UUID) (bool, error)
	IsMember(ctx context.Context, crewID, senderID uuid.UUID) (bool, error)
}

type crewRepository struct {
	db *gorm.DB
}

// constructor
func NewCrewRepository(db *gorm.DB) CrewRepository {
	return &crewRepository{db: db}
}

// create cerw
func (r *crewRepository) CreateCrew(crew *models.Crew) error {

	return r.db.Transaction(func(tx *gorm.DB) error {
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
func (r *crewRepository) FindForUser(userID string) ([]models.Crew, error) {
	var crews []models.Crew
	var user models.User

	// 1. First, find the user to start the association from.
	if err := r.db.First(&user, "id = ?", userID).Error; err != nil {
		return nil, err
	}

	// 2. Now, use Association to find all the 'Crews' linked to that user.
	err := r.db.Model(&user).Association("Crews").Find(&crews)
	return crews, err
}

// Delete the crews
func (r *crewRepository) DeleteCrewByID(ownerID uuid.UUID, crewID uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {

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

func (r *crewRepository) ExistByID(ctx context.Context, crewID uuid.UUID) (bool, error) {
	var count int64

	err := r.db.WithContext(ctx).Model(&models.Crew{}).Where("id=?", crewID).
		Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
func (r *crewRepository) IsMember(ctx context.Context, crewID, senderID uuid.UUID) (bool, error) {

	var count int64

	err := r.db.WithContext(ctx).
		Model(&models.CrewMember{}).Where("crew_id=? AND user_id=?", crewID, senderID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}
