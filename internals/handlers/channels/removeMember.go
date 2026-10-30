package channels

import (
	"chat-server/internals/middleware"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *ChannelHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()
	ctx := r.Context()

	// 1.Get the Owner ID
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		http.Error(w, "Invalid user", http.StatusUnauthorized)
		return
	}

	//2.Parse channelID from URL
	channelIDParam := chi.URLParam(r, "channelID")
	channelID, err := uuid.Parse(channelIDParam)
	if err != nil {
		http.Error(w, "Invalid channel ID", http.StatusBadRequest)
		return
	}

	//3.Parse memberID from URL
	memberIDParam := chi.URLParam(r, "userID")
	memberID, err := uuid.Parse(memberIDParam)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// // 4.check body for more info
	// var request requests.UpdateChannelMember
	// if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
	// 	http.Error(w, "Invalid user ID", http.StatusBadRequest)
	// 	return
	// }

	//call the service
	err = h.channelService.RemoveMember(ctx, channelID, userID, memberID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
