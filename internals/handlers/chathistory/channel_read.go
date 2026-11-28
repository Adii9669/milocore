package chathistory

import (
	"context"
	"encoding/json"
	"net/http"

	"chat-server/internals/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ChannelReadStore interface {
	AdvanceChannelRead(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) error
}

func (r SummaryRepository) AdvanceChannelRead(ctx context.Context, user, channel, message uuid.UUID) error {
	return r.DB.WithContext(ctx).Exec(`
UPDATE channel_members AS cm
SET last_read_at = m.created_at, last_read_message_id = m.id
FROM messages AS m
WHERE cm.user_id = ? AND cm.channel_id = ?
AND m.id = ? AND m.channel_id = cm.channel_id
AND (cm.last_read_at IS NULL OR (cm.last_read_at, cm.last_read_message_id) < (m.created_at, m.id))
`, user, channel, message).Error
}

func UpdateChannelReadMarker(store ChannelReadStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := middleware.GetUserIDFromContext(r.Context())
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		channel, err := uuid.Parse(chi.URLParam(r, "channelId"))
		var input struct {
			MessageID uuid.UUID `json:"messageId"`
		}
		if err != nil || channel == uuid.Nil ||
			json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024)).Decode(&input) != nil || input.MessageID == uuid.Nil {
			http.Error(w, "invalid channel or message ID", http.StatusBadRequest)
			return
		}
		if err := store.AdvanceChannelRead(r.Context(), user, channel, input.MessageID); err != nil {
			http.Error(w, "could not update channel reads", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
