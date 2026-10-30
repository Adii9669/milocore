package channels

import (
	"chat-server/internals/middleware"
	"chat-server/internals/utils"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (h *ChannelHandler) DeleteChannel(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()
	ctx := r.Context()

	// 2. Parse user ID
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		http.Error(w, "Invalid User ID", http.StatusBadRequest)
		return
	}

	// 3. Extract channel ID using Gorilla Mux
	channelIDStr := chi.URLParam(r, "channelID")
	channelID, err := uuid.Parse(channelIDStr)
	if err != nil {
		http.Error(w, "Invalid Channel ID", http.StatusBadRequest)
		return
	}

	// 4. Delete the channel
	err = h.channelService.DeleteChannel(ctx, channelID, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Channel not found", http.StatusForbidden)
			return
		}

		fmt.Printf("Error while deleting: %v\n", err)
		http.Error(w, "Failed to delete channel", http.StatusInternalServerError)
		return
	}

	// 5. Success
	w.Header().Set("Content-Type", "application/json")
	utils.PrettyJSON(w, map[string]any{"success": true})
}
