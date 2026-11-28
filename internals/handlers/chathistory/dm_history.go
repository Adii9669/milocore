package chathistory

import (
	"chat-server/internals/middleware"
	"chat-server/internals/utils"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *Handler) GetDmHistory(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	otherUserIDStr := chi.URLParam(r, "userId")
	otherUserID, err := uuid.Parse(otherUserIDStr)
	if err != nil {
		http.Error(w, "Invalid UserID", http.StatusBadRequest)
		return
	}

	authUserID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	limit, cursor, err := historyQuery(r)
	if err != nil {
		http.Error(w, "invalid page cursor", http.StatusBadRequest)
		return
	}

	messages, err := h.service.DmHistory(
		ctx,
		authUserID,
		otherUserID,
		limit,
		cursor,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	hasMore := len(messages) == limit
	nextCursor := pageCursor(messages, hasMore, r.URL.Query().Get("cursorFormat") == "stable")
	w.Header().Set("Content-Type", "application/json")
	utils.PrettyJSON(w, PaginatedMessageResponse{
		Message:    messages,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	})

}
