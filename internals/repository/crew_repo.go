package repository

import (
	"chat-server/internals/db/models"
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

var (
	ErrMemberNotFound = errors.New("member not found")
	ErrCrewNotFound   = errors.New("crew not found")
)

func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

type CrewRepository interface {

	//crew Methods
	CreateCrew(ctx context.Context, db *gorm.DB, crew *models.Crew) error
	RenameCrew(ctx context.Context, crewID uuid.UUID, name string) error
	DeleteCrew(ctx context.Context, crewID uuid.UUID) error
	IsMember(ctx context.Context, crewID, senderID uuid.UUID) (bool, error)

	//searching Methods
	FindByID(ctx context.Context, crewID uuid.UUID) (*models.Crew, error)
	FindForUser(ctx context.Context, userID uuid.UUID) ([]models.Crew, error)
	ExistByID(ctx context.Context, crewID uuid.UUID) (bool, error)

	//Members Methods

	CreateMember(ctx context.Context, db *gorm.DB, member *models.CrewMember) error
	FindMember(ctx context.Context, crewID, memberID uuid.UUID) (*models.CrewMember, error)
	RemoveMember(ctx context.Context, crewID, userID uuid.UUID) error
	UpdateRole(ctx context.Context, crewID, userID uuid.UUID, role models.CrewRole) error
	FindMembersByCrewID(ctx context.Context, crewID uuid.UUID) ([]models.CrewMember, error)
	ExistsMember(ctx context.Context, crewID, userID uuid.UUID) (bool, error)
}

type crewRepository struct {
	db *gorm.DB
}

// constructor
func NewCrewRepository(db *gorm.DB) CrewRepository {
	return &crewRepository{db: db}
}

// create cerw
func (r *crewRepository) CreateCrew(ctx context.Context, db *gorm.DB, crew *models.Crew) error {
	return db.WithContext(ctx).Create(crew).Error
}

// Delete the crews
func (r *crewRepository) DeleteCrew(ctx context.Context, crewID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Delete(&models.Crew{}, "id = ?", crewID)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrCrewNotFound
	}

	return nil
}

func (r *crewRepository) RenameCrew(ctx context.Context, crewID uuid.UUID, name string) error {
	result := r.db.WithContext(ctx).
		Model(&models.Crew{}).
		Where("id = ?", crewID).
		Update("name", name)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrCrewNotFound
	}
	return nil
}

// for FindForUser
func (r *crewRepository) FindForUser(
	ctx context.Context,
	userID uuid.UUID,
) ([]models.Crew, error) {

	var crews []models.Crew

	err := r.db.WithContext(ctx).
		Joins("JOIN crew_members ON crew_members.crew_id = crews.id").
		Where("crew_members.user_id = ?", userID).
		Find(&crews).Error

	if err != nil {
		return nil, err
	}

	return crews, nil
}

func (r *crewRepository) ExistByID(ctx context.Context, crewID uuid.UUID) (bool, error) {
	var count int64

	err := r.db.WithContext(ctx).WithContext(ctx).Model(&models.Crew{}).Where("id=?", crewID).
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

func (r *crewRepository) FindByID(
	ctx context.Context,
	crewID uuid.UUID,
) (*models.Crew, error) {

	var crew models.Crew

	err := r.db.WithContext(ctx).
		Where("id = ?", crewID).
		First(&crew).Error

	if err != nil {
		return nil, err
	}

	return &crew, nil
}

// Member
func (r *crewRepository) AddMember(ctx context.Context, member models.CrewMember) error {
	return r.db.WithContext(ctx).Create(&member).Error
}

// find
func (r *crewRepository) FindMember(
	ctx context.Context,
	crewID, memberID uuid.UUID,
) (*models.CrewMember, error) {

	var member models.CrewMember

	err := r.db.WithContext(ctx).
		Where("crew_id = ? AND user_id = ?", crewID, memberID).
		First(&member).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &member, nil
}

func (r *crewRepository) ExistsMember(ctx context.Context, crewID, userID uuid.UUID) (bool, error) {

	var count int64

	err := r.db.WithContext(ctx).
		Model(&models.CrewMember{}).
		Where("crew_id = ? AND user_id = ?", crewID, userID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// remove
func (r *crewRepository) RemoveMember(ctx context.Context, crewID, userID uuid.UUID) error {

	result := r.db.WithContext(ctx).
		Where("crew_id = ? AND user_id = ?", crewID, userID).
		Delete(&models.CrewMember{})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrMemberNotFound
	}
	return nil
}

// update role
func (r *crewRepository) UpdateRole(ctx context.Context, crewID, userID uuid.UUID, role models.CrewRole) error {

	result := r.db.WithContext(ctx).
		Model(&models.CrewMember{}).
		Where("crew_id = ? AND user_id = ?", crewID, userID).
		Update("role", role)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrCrewNotFound
	}

	return nil
}

func (r *crewRepository) CreateMember(ctx context.Context, db *gorm.DB, member *models.CrewMember) error {
	return db.WithContext(ctx).Create(member).Error
}

func (r *crewRepository) FindMembersByCrewID(ctx context.Context, crewID uuid.UUID) ([]models.CrewMember, error) {

	var members []models.CrewMember

	err := r.db.WithContext(ctx).
		Where("crew_id = ?", crewID).
		Find(&members).Error

	if err != nil {
		return nil, err
	}

	return members, nil

}
