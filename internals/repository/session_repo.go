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
	FindByTokenHash(ctx context.Context, hash string) (*models.Session, error)
	DeleteSession(ctx context.Context, userID, sessionID uuid.UUID) error
	DeleteExpired(ctx context.Context) error
}

type sessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) SessionRepository {
	return &sessionRepository{db: db}
}

func (r *sessionRepository) Create(ctx context.Context, session *models.Session) error {
	return r.db.WithContext(ctx).Create(session).Error
}

// FindByTokenHash looks a session up by the hashed refresh token. This lets
// the refresh endpoint identify the user even after the access token cookie
// has expired and disappeared from the browser.
func (r *sessionRepository) FindByTokenHash(ctx context.Context, hash string) (*models.Session, error) {
	var session models.Session
	err := r.db.WithContext(ctx).
		Where("refresh_token_hash = ? AND expires_at > ?", hash, time.Now()).
		First(&session).Error
	return &session, err
}

func (r *sessionRepository) DeleteSession(ctx context.Context, userID, sessionID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", sessionID, userID).
		Delete(&models.Session{}).Error
}

func (r *sessionRepository) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&models.Session{}).Error
}
