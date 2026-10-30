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
	ErrMemberNotFound  = errors.New("member not found")
	ErrChannelNotFound = errors.New("channel not found")
)

func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

type ChannelRepository interface {

	//channel Methods
	CreateChannel(ctx context.Context, db *gorm.DB, channel *models.Channel) error
	RenameChannel(ctx context.Context, channelID uuid.UUID, name string) error
	DeleteChannel(ctx context.Context, channelID uuid.UUID) error
	IsMember(ctx context.Context, channelID, senderID uuid.UUID) (bool, error)

	//searching Methods
	FindByID(ctx context.Context, channelID uuid.UUID) (*models.Channel, error)
	FindForUser(ctx context.Context, userID uuid.UUID) ([]models.Channel, error)
	ExistByID(ctx context.Context, channelID uuid.UUID) (bool, error)

	//Members Methods

	CreateMember(ctx context.Context, db *gorm.DB, member *models.ChannelMember) error
	FindMember(ctx context.Context, channelID, memberID uuid.UUID) (*models.ChannelMember, error)
	RemoveMember(ctx context.Context, channelID, userID uuid.UUID) error
	UpdateRole(ctx context.Context, channelID, userID uuid.UUID, role models.ChannelRole) error
	FindMembersByChannelID(ctx context.Context, channelID uuid.UUID) ([]models.ChannelMember, error)
	ExistsMember(ctx context.Context, channelID, userID uuid.UUID) (bool, error)
}

type channelRepository struct {
	db *gorm.DB
}

// constructor
func NewChannelRepository(db *gorm.DB) ChannelRepository {
	return &channelRepository{db: db}
}

// create channel
func (r *channelRepository) CreateChannel(ctx context.Context, db *gorm.DB, channel *models.Channel) error {
	return db.WithContext(ctx).Create(channel).Error
}

// Delete the channels
func (r *channelRepository) DeleteChannel(ctx context.Context, channelID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Delete(&models.Channel{}, "id = ?", channelID)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrChannelNotFound
	}

	return nil
}

func (r *channelRepository) RenameChannel(ctx context.Context, channelID uuid.UUID, name string) error {
	result := r.db.WithContext(ctx).
		Model(&models.Channel{}).
		Where("id = ?", channelID).
		Update("name", name)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrChannelNotFound
	}
	return nil
}

// for FindForUser
func (r *channelRepository) FindForUser(
	ctx context.Context,
	userID uuid.UUID,
) ([]models.Channel, error) {

	var channels []models.Channel

	err := r.db.WithContext(ctx).
		Joins("JOIN channel_members ON channel_members.channel_id = channels.id").
		Where("channel_members.user_id = ?", userID).
		Find(&channels).Error

	if err != nil {
		return nil, err
	}

	return channels, nil
}

func (r *channelRepository) ExistByID(ctx context.Context, channelID uuid.UUID) (bool, error) {
	var count int64

	err := r.db.WithContext(ctx).WithContext(ctx).Model(&models.Channel{}).Where("id=?", channelID).
		Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
func (r *channelRepository) IsMember(ctx context.Context, channelID, senderID uuid.UUID) (bool, error) {

	var count int64

	err := r.db.WithContext(ctx).
		Model(&models.ChannelMember{}).Where("channel_id=? AND user_id=?", channelID, senderID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *channelRepository) FindByID(
	ctx context.Context,
	channelID uuid.UUID,
) (*models.Channel, error) {

	var channel models.Channel

	err := r.db.WithContext(ctx).
		Where("id = ?", channelID).
		First(&channel).Error

	if err != nil {
		return nil, err
	}

	return &channel, nil
}

// Member
func (r *channelRepository) AddMember(ctx context.Context, member models.ChannelMember) error {
	return r.db.WithContext(ctx).Create(&member).Error
}

// find
func (r *channelRepository) FindMember(
	ctx context.Context,
	channelID, memberID uuid.UUID,
) (*models.ChannelMember, error) {

	var member models.ChannelMember

	err := r.db.WithContext(ctx).
		Where("channel_id = ? AND user_id = ?", channelID, memberID).
		First(&member).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &member, nil
}

func (r *channelRepository) ExistsMember(ctx context.Context, channelID, userID uuid.UUID) (bool, error) {

	var count int64

	err := r.db.WithContext(ctx).
		Model(&models.ChannelMember{}).
		Where("channel_id = ? AND user_id = ?", channelID, userID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// remove
func (r *channelRepository) RemoveMember(ctx context.Context, channelID, userID uuid.UUID) error {

	result := r.db.WithContext(ctx).
		Where("channel_id = ? AND user_id = ?", channelID, userID).
		Delete(&models.ChannelMember{})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrMemberNotFound
	}
	return nil
}

// update role
func (r *channelRepository) UpdateRole(ctx context.Context, channelID, userID uuid.UUID, role models.ChannelRole) error {

	result := r.db.WithContext(ctx).
		Model(&models.ChannelMember{}).
		Where("channel_id = ? AND user_id = ?", channelID, userID).
		Update("role", role)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrChannelNotFound
	}

	return nil
}

func (r *channelRepository) CreateMember(ctx context.Context, db *gorm.DB, member *models.ChannelMember) error {
	return db.WithContext(ctx).Create(member).Error
}

func (r *channelRepository) FindMembersByChannelID(ctx context.Context, channelID uuid.UUID) ([]models.ChannelMember, error) {

	var members []models.ChannelMember

	err := r.db.WithContext(ctx).
		Where("channel_id = ?", channelID).
		Find(&members).Error

	if err != nil {
		return nil, err
	}

	return members, nil

}
