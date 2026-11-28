package chathistory

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"chat-server/internals/middleware"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ConversationSummary struct {
	ID            uuid.UUID  `json:"id"`
	Name          string     `json:"name"`
	Kind          string     `json:"kind"`
	LastMessageAt *time.Time `json:"lastMessageAt,omitempty"`
	Preview       *string    `json:"preview,omitempty"`
	Unread        *int64     `json:"unread,omitempty"`
}

type SummaryStore interface {
	List(context.Context, uuid.UUID) ([]ConversationSummary, error)
}

type SummaryRepository struct{ DB *gorm.DB }

func (r SummaryRepository) List(ctx context.Context, user uuid.UUID) ([]ConversationSummary, error) {
	rows := make([]ConversationSummary, 0)
	err := r.DB.WithContext(ctx).Raw(`
WITH my_channels AS (
	SELECT c.id, c.name, cm.joined_at, cm.last_read_at, cm.last_read_message_id
	FROM channels c JOIN channel_members cm ON cm.channel_id = c.id
	WHERE cm.user_id = @user
), contacts AS (
	SELECT u.id, u.name
	FROM friends f
	JOIN users u ON u.id = CASE WHEN f.user_low_id = @user THEN f.user_high_id ELSE f.user_low_id END
	WHERE f.status = 'accepted' AND (f.user_low_id = @user OR f.user_high_id = @user)
), summaries AS (
	SELECT c.id, c.name, 'channel' AS kind, latest.created_at AS last_message_at, latest.preview, incoming.total AS unread
	FROM my_channels c
	LEFT JOIN LATERAL (
		SELECT m.created_at, LEFT(COALESCE(NULLIF(m.content, ''), m.file_name, ''), 280) AS preview
		FROM messages m
		WHERE m.channel_id = c.id AND NOT EXISTS (
			SELECT 1 FROM message_user_deletions d WHERE d.message_id = m.id AND d.user_id = @user
		)
		ORDER BY m.created_at DESC, m.id DESC LIMIT 1
	) latest ON TRUE
	CROSS JOIN LATERAL (
		SELECT count(*) AS total FROM messages m
		WHERE m.channel_id = c.id AND m.sender_id <> @user
		AND m.created_at >= c.joined_at
		AND (c.last_read_at IS NULL OR (m.created_at, m.id) > (c.last_read_at, c.last_read_message_id))
		AND NOT EXISTS (
			SELECT 1 FROM message_user_deletions d WHERE d.message_id = m.id AND d.user_id = @user
		)
	) incoming
	UNION ALL
	SELECT c.id, c.name, 'dm', latest.created_at, latest.preview, incoming.total
	FROM contacts c
	LEFT JOIN LATERAL (
		SELECT m.created_at, LEFT(COALESCE(NULLIF(m.content, ''), m.file_name, ''), 280) AS preview
		FROM messages m
		WHERE m.channel_id IS NULL AND (
			(m.sender_id = @user AND m.receiver_id = c.id)
			OR (m.sender_id = c.id AND m.receiver_id = @user)
		) AND NOT EXISTS (
			SELECT 1 FROM message_user_deletions d WHERE d.message_id = m.id AND d.user_id = @user
		)
		ORDER BY m.created_at DESC, m.id DESC LIMIT 1
	) latest ON TRUE
	CROSS JOIN LATERAL (
		SELECT count(*) AS total FROM messages m
		WHERE m.sender_id = c.id AND m.receiver_id = @user AND m.channel_id IS NULL AND m.read = false
		AND NOT EXISTS (
			SELECT 1 FROM message_user_deletions d WHERE d.message_id = m.id AND d.user_id = @user
		)
	) incoming
)
SELECT id, name, kind, last_message_at, preview, unread FROM summaries
ORDER BY last_message_at DESC NULLS LAST, lower(name), kind, id
`, map[string]any{"user": user}).Scan(&rows).Error
	return rows, err
}

func ListSummaries(store SummaryStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := middleware.GetUserIDFromContext(r.Context())
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		summaries, err := store.List(r.Context(), user)
		if err != nil {
			http.Error(w, "conversation summaries unavailable", http.StatusServiceUnavailable)
			return
		}
		if summaries == nil {
			summaries = []ConversationSummary{}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(summaries)
	}
}
