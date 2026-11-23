package repository

import (
	"chat-server/internals/db/models"
	"context"
	"errors"
	"slices"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ErrChannelAccessDenied is returned when a user asks for history from a
// channel they do not belong to. The service layer translates this into a
// 403 instead of leaking whether the channel exists.
var ErrChannelAccessDenied = errors.New("channel access denied")

type MessageRepository interface {
	SaveMessage(context.Context, *models.Message) error
	FindMessageByID(context.Context, uuid.UUID) (*models.Message, error)

	GetDmMessageHistory(
		ctx context.Context,
		userA uuid.UUID,
		userB uuid.UUID,
		limit int,
		cursor *time.Time,
	) ([]models.Message, error)
	GetChannelMessageHistory(
		ctx context.Context,
		channelID uuid.UUID,
		limit int,
		cursor *time.Time,
	) ([]models.Message, error)
	MarkDelivered(ctx context.Context, messageID uuid.UUID) error
	MarkRead(ctx context.Context, userID, otherUserID uuid.UUID) error
}

// MessageDeletionRepository provides persistent per-user hiding and removal.
type MessageDeletionRepository interface {
	HideMessageForUser(context.Context, uuid.UUID, uuid.UUID) error
	IsMessageHiddenForUser(context.Context, uuid.UUID, uuid.UUID) (bool, error)
	DeleteMessage(context.Context, uuid.UUID) error
	GetDmMessageHistoryForUser(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, int, *time.Time) ([]models.Message, error)
	GetChannelMessageHistoryForUser(context.Context, uuid.UUID, uuid.UUID, int, *time.Time) ([]models.Message, error)
}

type messageRepository struct {
	db *gorm.DB
}

// constructor
func NewMessageRepository(db *gorm.DB) MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) FindMessageByID(ctx context.Context, id uuid.UUID) (*models.Message, error) {
	var message models.Message
	err := r.db.WithContext(ctx).Preload("Sender").First(&message, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}

// GetChannelMessageHistory fetches group chat history without a per-user
// deletion filter. It is kept for repositories/tests that do not implement
// MessageDeletionRepository.
func (r *messageRepository) GetChannelMessageHistory(
	ctx context.Context,
	channelID uuid.UUID,
	limit int,
	cursor *time.Time,
) ([]models.Message, error) {

	var messages []models.Message

	// Joins("Sender") makes GORM load the sender in the same SQL statement.
	// The old Preload("Sender") did a second round trip for every page.
	query := r.db.
		WithContext(ctx).
		Joins("Sender").
		Where("messages.channel_id = ?", channelID)

	if cursor != nil {
		query = query.Where("messages.created_at < ?", *cursor)
	}

	err := query.
		Order("messages.created_at DESC").
		Limit(limit).
		Find(&messages).Error
	if err != nil {
		return nil, err
	}

	// Reverse in memory so frontend gets ASC order.
	slices.Reverse(messages)

	return messages, nil
}

// saves the messages to the database
// using the ctx for leting is the user is disconnects.
// if client disconnects if WS closes if request is canceled
func (r *messageRepository) SaveMessage(ctx context.Context, msg *models.Message) error {
	return r.db.WithContext(ctx).Create(msg).Error
}

func (r *messageRepository) GetDmMessageHistory(
	ctx context.Context,
	userA uuid.UUID,
	userB uuid.UUID,
	limit int,
	cursor *time.Time,
) ([]models.Message, error) {
	var messages []models.Message

	query := r.db.
		WithContext(ctx).
		Joins("Sender").
		Where(
			"((messages.sender_id = ? AND messages.receiver_id = ?) OR (messages.sender_id = ? AND messages.receiver_id = ?))",
			userA, userB,
			userB, userA,
		)

	if cursor != nil {
		query = query.Where("messages.created_at < ?", *cursor)
	}

	err := query.
		Order("messages.created_at DESC").
		Limit(limit).
		Find(&messages).Error

	if err != nil {
		return nil, err
	}

	// Reverse in memory so frontend gets ASC order
	slices.Reverse(messages)

	return messages, nil
}

func (r *messageRepository) MarkDelivered(ctx context.Context, messageID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&models.Message{}).
		Where("id = ?", messageID).
		Update("delivered", true).Error
}

func (r *messageRepository) MarkRead(ctx context.Context, userID, otherUserID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&models.Message{}).
		Where("sender_id = ? AND receiver_id = ?", otherUserID, userID).
		Where("read = ?", false).
		Update("read", true).Error
}

func (r *messageRepository) HideMessageForUser(ctx context.Context, messageID, userID uuid.UUID) error {
	hidden := models.MessageUserDeletion{MessageID: messageID, UserID: userID}
	return r.db.WithContext(ctx).FirstOrCreate(&hidden, hidden).Error
}

func (r *messageRepository) IsMessageHiddenForUser(ctx context.Context, messageID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.MessageUserDeletion{}).Where("message_id = ? AND user_id = ?", messageID, userID).Count(&count).Error
	return count > 0, err
}

func (r *messageRepository) DeleteMessage(ctx context.Context, messageID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("message_id = ?", messageID).Delete(&models.MessageUserDeletion{}).Error; err != nil {
			return err
		}
		return tx.Unscoped().Delete(&models.Message{}, "id = ?", messageID).Error
	})
}

func (r *messageRepository) GetDmMessageHistoryForUser(ctx context.Context, userA, userB, currentUser uuid.UUID, limit int, cursor *time.Time) ([]models.Message, error) {
	var messages []models.Message
	query := r.db.WithContext(ctx).Joins("Sender").Where(
		"((messages.sender_id = ? AND messages.receiver_id = ?) OR (messages.sender_id = ? AND messages.receiver_id = ?)) AND NOT EXISTS (SELECT 1 FROM message_user_deletions mud WHERE mud.message_id = messages.id AND mud.user_id = ?)",
		userA, userB, userB, userA, currentUser,
	)
	if cursor != nil {
		query = query.Where("messages.created_at < ?", *cursor)
	}
	err := query.Order("messages.created_at DESC").Limit(limit).Find(&messages).Error
	if err != nil {
		return nil, err
	}
	slices.Reverse(messages)
	return messages, nil
}

func (r *messageRepository) GetChannelMessageHistoryForUser(ctx context.Context, channelID, currentUser uuid.UUID, limit int, cursor *time.Time) ([]models.Message, error) {
	var messages []models.Message

	// One SELECT:
	//   - messages for the channel
	//   - sender joined through users
	//   - membership enforced by joining channel_members
	//   - per-user deleted messages excluded with NOT EXISTS
	query := r.db.WithContext(ctx).
		Joins("Sender").
		Joins("JOIN channel_members AS cm ON cm.channel_id = messages.channel_id AND cm.user_id = ?", currentUser).
		Where("messages.channel_id = ?", channelID).
		Where("NOT EXISTS (SELECT 1 FROM message_user_deletions mud WHERE mud.message_id = messages.id AND mud.user_id = ?)", currentUser)

	if cursor != nil {
		query = query.Where("messages.created_at < ?", *cursor)
	}

	err := query.Order("messages.created_at DESC").Limit(limit).Find(&messages).Error
	if err != nil {
		return nil, err
	}

	// A zero-row result can mean either "empty channel" or "not a member".
	// Only pay for the membership check when the page is empty. Normal pages
	// with messages already proved membership through the JOIN above.
	if len(messages) == 0 {
		var membershipCount int64
		if err := r.db.WithContext(ctx).
			Model(&models.ChannelMember{}).
			Where("channel_id = ? AND user_id = ?", channelID, currentUser).
			Count(&membershipCount).Error; err != nil {
			return nil, err
		}
		if membershipCount == 0 {
			return nil, ErrChannelAccessDenied
		}
	}

	slices.Reverse(messages)
	return messages, nil
}
