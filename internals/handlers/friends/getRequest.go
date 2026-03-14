package friends

import (
	"chat-server/internals/middleware"
	"chat-server/internals/utils"
	"net/http"
)

func (h *FriendHandler) GetRequests(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		http.Error(w, "Unauthozied", http.StatusUnauthorized)
		return
	}

	reqType := r.URL.Query().Get("type")

	requests, err := h.friendService.GetRequests(ctx, userID, reqType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	utils.PrettyJSON(w, map[string]any{
		"message": "success",
		"send":    requests,
	})

}
