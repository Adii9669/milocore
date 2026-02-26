package repository

import (
	"chat-server/internals/db/models"
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FriendRepository interface {
	Create(ctx context.Context, friend *models.Friend) error
	FindBYPair(ctx context.Context, low, high uuid.UUID) (*models.Friend, error)
	Update(ctx context.Context, friend *models.Friend) error
	GetAcceptedByUser(ctx context.Context, userID uuid.UUID) ([]models.Friend, error)
	Delete(ctx context.Context, low, high uuid.UUID) error
	GetRelation(ctx context.Context, low, high uuid.UUID) (*models.Friend, error)
}

type friendRepository struct {
	db *gorm.DB
}

func NewFriendRepository(db *gorm.DB) FriendRepository {
	return &friendRepository{db: db}
}

func (r *friendRepository) Create(ctx context.Context, friend *models.Friend) error {
	return r.db.WithContext(ctx).Create(friend).Error
}

func (r *friendRepository) FindBYPair(ctx context.Context, low, high uuid.UUID) (*models.Friend, error) {

	var friend models.Friend

	err := r.db.WithContext(ctx).Where("user_low_id = ? AND user_high_id = ?", low, high).First(&friend).Error

	if err != nil {
		return nil, err
	}
	return &friend, nil
}

func (r *friendRepository) Update(
	ctx context.Context,
	friend *models.Friend,
) error {
	return r.db.WithContext(ctx).Save(friend).Error
}

func (r *friendRepository) GetAcceptedByUser(
	ctx context.Context,
	userID uuid.UUID,
) ([]models.Friend, error) {

	var friends []models.Friend

	err := r.db.WithContext(ctx).
		Where(
			"(user_low_id = ? OR user_high_id = ?) AND status = ?",
			userID, userID, "accepted",
		).
		Find(&friends).Error

	return friends, err
}

func (r *friendRepository) Delete(
	ctx context.Context,
	low uuid.UUID,
	high uuid.UUID,
) error {
	return r.db.WithContext(ctx).
		Where("user_low_id = ? AND user_high_id = ?", low, high).
		Delete(&models.Friend{}).
		Error
}

func (r *friendRepository) GetRelation(
	ctx context.Context,
	low, high uuid.UUID,
) (*models.Friend, error) {

	var relation models.Friend

	err := r.db.WithContext(ctx).
		Where("user_low_id = ? AND user_high_id = ?", low, high).
		First(&relation).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &relation, nil
}
