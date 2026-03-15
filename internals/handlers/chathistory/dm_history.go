package chathistory

import (
	"chat-server/internals/middleware"
	"chat-server/internals/utils"
	"log"
	"net/http"
	"strconv"
	"time"

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

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	var cursor *time.Time
	if cursorStr := r.URL.Query().Get("cursor"); cursorStr != "" {
		parsed, err := time.Parse(time.RFC3339, cursorStr)
		if err == nil {
			cursor = &parsed
		}
	}

	// log.Printf("otheruser count %d", len(otherUserID))
	// log.Printf("auth count %d", len(authUserID))
	// log.Printf("userA %s", otherUserID)
	// log.Printf("userB %s", authUserID)

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

	var nextCursor *time.Time
	hasMore := false

	if len(messages) == limit {
		hasMore = true
		oldest := messages[len(messages)-1]
		nextCursor = &oldest.CreatedAt
	}
	w.Header().Set("Content-Type", "application/json")
	utils.PrettyJSON(w, PaginatedMessageResponse{
		Message:    messages,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	})

}
