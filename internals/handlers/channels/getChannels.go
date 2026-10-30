package channels

import (
	"chat-server/internals/middleware"
	"chat-server/internals/utils"

	// "log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *ChannelHandler) GetChannels(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()
	//1. always get the claims first authentication from middleware
	ctx := r.Context()
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	response, err := h.channelService.GetChannels(ctx, userID)
	if err != nil {
		http.Error(w, "Failed to retrive the Channel", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	utils.PrettyJSON(w, map[string]any{
		"message":  "Channels fetched successfully",
		"channels": response,
	})
}

func (h *ChannelHandler) GetMembers(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()
	//1. always get the claims first authentication from middleware
	ctx := r.Context()
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	channelIDStr := chi.URLParam(r, "channelID")
	channelID, err := uuid.Parse(channelIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// var req requests.GetMembers
	// if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
	// 	http.Error(w, "Invalid user ID", http.StatusBadRequest)
	// 	return
	// }

	response, err := h.channelService.GetMembers(ctx, channelID, userID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	utils.PrettyJSON(w, map[string]any{
		"members": response,
	})
}
