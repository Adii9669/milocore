package repository

import (
	"chat-server/internals/db"
	"chat-server/internals/db/models"

	"github.com/google/uuid"
)

// this is the contract for using the database operation
type UserRepository interface {
	FindByID(id uuid.UUID) (*models.User, error)
	FindByEmail(email string) (*models.User, error)
	FindBYName(name string) (*models.User, error)
	Create(user *models.User) error
}

// GormUserRepository is the GORM implementation of our repository.
type userRepository struct{}

// constructor
func NewUserRepository() UserRepository {
	return &userRepository{}
}

// (id)FindByID retrieves a user by their ID from the database.
func (r *userRepository) FindByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := db.DB.First(&user, "id=?", id).Error
	return &user, err
}

// name
func (r *userRepository) FindBYName(name string) (*models.User, error) {
	var user models.User
	err := db.DB.First(&user, "name=?", name).Error
	return &user, err
}

// FindByEmail retrieves a user by their email from the database.
func (r *userRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := db.DB.First(&user, "email = ?", email).Error
	return &user, err
}

// Create saves a new user record to the database.
func (r *userRepository) Create(user *models.User) error {
	return db.DB.Create(user).Error
}
