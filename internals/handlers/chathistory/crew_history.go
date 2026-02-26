package chathistory

import (
	"chat-server/internals/middleware"
	"chat-server/internals/transport/dto"
	"chat-server/internals/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type PaginatedMessageResponse struct {
	Message    []dto.MessageResponse `json:"messages"`
	NextCursor *time.Time            `json:"nextCursor"`
	HasMore    bool                  `json:"hasMore"`
}

func (h *Handler) GetCrewHistory(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		http.Error(w, "Invalid User ID", http.StatusUnauthorized)
		return
	}

	crewIDstr := chi.URLParam(r, "crewId")
	crewID, err := uuid.Parse(crewIDstr)
	if err != nil {
		http.Error(w, "Invalid crew ID", http.StatusBadRequest)
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

	messages, err := h.service.CrewHistory(ctx, crewID, userID.String(), limit, cursor)
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
	utils.PrettyJSON(w,
		PaginatedMessageResponse{
			Message:    messages,
			NextCursor: nextCursor,
			HasMore:    hasMore,
		})
}
