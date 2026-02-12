package repository

import (
	"chat-server/internals/db/models"
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// this is the contract for using the database operation
type UserRepository interface {
	FindByID(id uuid.UUID) (*models.User, error)
	FindByEmail(email string) (*models.User, error)
	FindBYName(name string) (*models.User, error)
	Create(user *models.User) error
	ExistByID(ctx context.Context, userID uuid.UUID) (bool, error)
}

// GormUserRepository is the GORM implementation of our repository.
type userRepository struct {
	db *gorm.DB
}

// constructor
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// (id)FindByID retrieves a user by their ID from the database.
func (r *userRepository) FindByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, "id=?", id).Error
	return &user, err
}

// name
func (r *userRepository) FindBYName(name string) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, "name=?", name).Error
	return &user, err
}

// FindByEmail retrieves a user by their email from the database.
func (r *userRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, "email = ?", email).Error
	return &user, err
}

// Create saves a new user record to the database.
func (r *userRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

// ExistByID
func (r *userRepository) ExistByID(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64

	err := r.db.WithContext(ctx).Model(&models.User{}).Where("id=?", id).Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}
