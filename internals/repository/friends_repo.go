package repository

import (
	"chat-server/internals/db"
	"chat-server/internals/db/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FriendRepository interface {
	SendRequest(userID, friendID uuid.UUID) error
	AcceptRequest(rel *models.Friends) error
	FindRelation(userID, freindID uuid.UUID) (*models.Friends, error)
	CreateReverseAccepted(userID, friendID uuid.UUID) error
	GetFriends(userID uuid.UUID) ([]models.Friends, error)
	AcceptAndCreateReverse(rel *models.Friends, accepterID uuid.UUID) error
}

type friendRepository struct{}

func NewFriendRepository() FriendRepository {
	return &friendRepository{}
}

func (r *friendRepository) SendRequest(userID, friendID uuid.UUID) error {

	friend := models.Friends{
		ID:       uuid.New(),
		UserID:   userID,
		FriendID: friendID,
		Status:   "pending",
	}
	return db.DB.Create(&friend).Error
}

func (r *friendRepository) FindRelation(userID, frndID uuid.UUID) (*models.Friends, error) {

	var rel models.Friends

	err := db.DB.Where("user_id = ? AND friend_id = ? ", userID, frndID).First(&rel).Error

	return &rel, err
}

func (r *friendRepository) AcceptRequest(rel *models.Friends) error {
	/*
		We modify the struct passed into the function.
		This updates the existing database row.
	*/
	rel.Status = "accepted"
	return db.DB.Save(rel).Error
}

func (r *friendRepository) CreateReverseAccepted(userID, friendID uuid.UUID) error {

	rev := models.Friends{
		ID:       uuid.New(),
		UserID:   userID,   // the accepter
		FriendID: friendID, // the sender originally
		Status:   "accepted",
	}

	return db.DB.Create(&rev).Error
}

// -------------------------------------------------------------
func (r *friendRepository) GetFriends(userID uuid.UUID) ([]models.Friends, error) {
	var friends []models.Friends

	/*
		This returns rows like:
		userID → friendID
		where status = 'accepted'

		This gives all users that THIS user is friends with.
	*/

	err := db.DB.
		Where("user_id = ? AND status = 'accepted'", userID).
		Find(&friends).Error

	return friends, err
}

func (r *friendRepository) AcceptAndCreateReverse(rel *models.Friends, accepterID uuid.UUID) error {
	// run everything inside a transaction
	return db.DB.Transaction(func(tx *gorm.DB) error {

		// 1) Update the existing relation (A -> B) to accepted.
		rel.Status = "accepted"

		if err := tx.Save(rel).Error; err != nil {
			// returning an error rolls back the transaction
			return err
		}

		// 2) Create the reverse accepted relation (B -> A)
		rev := models.Friends{
			ID:       uuid.New(),
			UserID:   accepterID, // B
			FriendID: rel.UserID, // A
			Status:   "accepted",
		}

		if err := tx.Create(&rev).Error; err != nil {
			// If a UNIQUE constraint is in place this could fail with a duplicate-key error.
			// For simplicity we return the error here and let the caller decide how to respond.
			// Optionally you can detect the unique-violation error (SQLSTATE 23505) and ignore it.
			return err
		}

		// commit (return nil)
		return nil
	})
}
