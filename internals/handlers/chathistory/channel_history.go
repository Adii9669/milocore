package chathistory

import (
	"chat-server/internals/middleware"
	"chat-server/internals/services"
	"chat-server/internals/transport/dto"
	"chat-server/internals/utils"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type PaginatedMessageResponse struct {
	Message    []dto.MessageResponse `json:"messages"`
	NextCursor *string               `json:"nextCursor"`
	HasMore    bool                  `json:"hasMore"`
}

func (h *Handler) GetChannelHistory(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		http.Error(w, "Invalid User ID", http.StatusUnauthorized)
		return
	}

	channelIDstr := chi.URLParam(r, "channelId")
	channelID, err := uuid.Parse(channelIDstr)
	if err != nil {
		http.Error(w, "Invalid channel ID", http.StatusBadRequest)
		return
	}

	limit, cursor, err := historyQuery(r)
	if err != nil {
		http.Error(w, "invalid page cursor", http.StatusBadRequest)
		return
	}

	messages, err := h.service.ChannelHistory(ctx, channelID, userID, limit, cursor)
	if err != nil {
		if errors.Is(err, services.ErrChatAccessDenied) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	hasMore := len(messages) == limit
	nextCursor := pageCursor(messages, hasMore, r.URL.Query().Get("cursorFormat") == "stable")
	w.Header().Set("Content-Type", "application/json")
	utils.PrettyJSON(w,
		PaginatedMessageResponse{
			Message:    messages,
			NextCursor: nextCursor,
			HasMore:    hasMore,
		})
}
