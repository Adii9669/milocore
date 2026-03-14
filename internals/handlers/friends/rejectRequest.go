package friends

import (
	"chat-server/internals/middleware"
	"chat-server/internals/utils"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *FriendHandler) RejectRequest(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		http.Error(w, "Invalid user id", http.StatusUnauthorized)
		return
	}

	//2.Take user id from the param
	friendIDStr := chi.URLParam(r, "friendID")
	friendID, err := uuid.Parse(friendIDStr)
	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}

	err = h.friendService.RejectRequest(ctx, userID, friendID)
	if err != nil {
		log.Println("Reject Request Failed", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	utils.PrettyJSON(w, map[string]string{
		"success": "true",
	})
}
