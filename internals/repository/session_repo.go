package repository

import (
	"chat-server/internals/db/models"
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SessionRepository interface {
	Create(ctx context.Context, session *models.Session) error
	FindByUserID(ctx context.Context, userID uuid.UUID) (*models.Session, error)
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
	DeleteExpired(ctx context.Context) error
}

type sessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) SessionRepository {
	return &sessionRepository{db: db}
}

func (r *sessionRepository) Create(ctx context.Context, session *models.Session) error {
	// delete any existing session for this user first — one session per user
	r.db.WithContext(ctx).Where("user_id = ?", session.UserID).Delete(&models.Session{})
	return r.db.WithContext(ctx).Create(session).Error
}

func (r *sessionRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*models.Session, error) {
	var session models.Session
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND expires_at > ?", userID, time.Now()).
		First(&session).Error
	return &session, err
}

func (r *sessionRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&models.Session{}).Error
}

func (r *sessionRepository) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&models.Session{}).Error
}
