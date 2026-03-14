package repository

import (
	"chat-server/internals/db/models"
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// this is the contract for using the database operation
type UserRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindBYName(ctx context.Context, name string) (*models.User, error)
	Create(ctx context.Context, user *models.User) error
	ExistByID(ctx context.Context, userID uuid.UUID) (bool, error)
	FindAll(ctx context.Context) ([]models.User, error)

	FindByIDs(ctx context.Context, ids []uuid.UUID) ([]models.User, error)
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
func (r *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {

	var user models.User
	result := r.db.WithContext(ctx).First(&user, "id=?", id)
	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

// name
func (r *userRepository) FindBYName(ctx context.Context, name string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).First(&user, "name=?", name).Error
	return &user, err
}

// FindByEmail retrieves a user by their email from the database.
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).First(&user, "email = ?", email).Error
	return &user, err
}

// Create saves a new user record to the database.
func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
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

func (r *userRepository) FindByIDs(
	ctx context.Context,
	ids []uuid.UUID,
) ([]models.User, error) {

	if len(ids) == 0 {
		return []models.User{}, nil
	}

	var users []models.User

	err := r.db.WithContext(ctx).
		Where("id IN ?", ids).
		Find(&users).Error

	return users, err
}

func (r *userRepository) FindAll(ctx context.Context) ([]models.User, error) {
	var users []models.User
	err := r.db.WithContext(ctx).Find(&users).Error
	return users, err
}
