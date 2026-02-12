package repository

import (
	"chat-server/internals/db/models"
	"chat-server/internals/transport/dto"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FriendRepository interface {
	SendRequest(userID, friendID uuid.UUID) error
	AcceptRequest(rel *models.Friends) error
	FindRelation(userID, freindID uuid.UUID) (*models.Friends, error)
	CreateReverseAccepted(userID, friendID uuid.UUID) error
	GetFriend(userID uuid.UUID) ([]dto.FriendResponse, error)
	AcceptAndCreateReverse(rel *models.Friends, accepterID uuid.UUID) error
}

type friendRepository struct {
	db *gorm.DB
}

func NewFriendRepository(db *gorm.DB) FriendRepository {
	return &friendRepository{db: db}
}

func (r *friendRepository) SendRequest(userID, friendID uuid.UUID) error {

	friend := models.Friends{
		ID:       uuid.New(),
		UserID:   userID,
		FriendID: friendID,
		Status:   "pending",
	}
	return r.db.Create(&friend).Error
}

func (r *friendRepository) FindRelation(userID, frndID uuid.UUID) (*models.Friends, error) {

	var rel models.Friends

	err := r.db.Where("user_id = ? AND friend_id = ? ", userID, frndID).First(&rel).Error

	return &rel, err
}

func (r *friendRepository) AcceptRequest(rel *models.Friends) error {
	/*
		We modify the struct passed into the function.
		This updates the existing database row.
	*/
	rel.Status = "accepted"
	return r.db.Save(rel).Error
}

func (r *friendRepository) CreateReverseAccepted(userID, friendID uuid.UUID) error {

	rev := models.Friends{
		ID:       uuid.New(),
		UserID:   userID,   // the accepter
		FriendID: friendID, // the sender originally
		Status:   "accepted",
	}

	return r.db.Create(&rev).Error
}

// -------------------------------------------------------------
func (r *friendRepository) GetFriend(userID uuid.UUID) ([]dto.FriendResponse, error) {
	var friends []dto.FriendResponse

	err := r.db.
		Table("friends").
		Select(`
			users.id AS id,
			users.name AS name,
			friends.status AS status,
			friends.created_at AS created_at
		`).
		Joins("JOIN users ON users.id = friends.friend_id").
		Where("friends.user_id = ? AND friends.status = 'accepted'", userID).
		Scan(&friends).Error
	fmt.Printf("%+v\n", friends)
	return friends, err
}

func (r *friendRepository) AcceptAndCreateReverse(rel *models.Friends, accepterID uuid.UUID) error {
	// run everything inside a transaction
	return r.db.Transaction(func(tx *gorm.DB) error {

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
