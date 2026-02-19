package friends

import (
	"chat-server/internals/transport/dto"
	"chat-server/internals/utils"
	"chat-server/middleware"
	"encoding/json"

	// "log"
	"net/http"

	"github.com/google/uuid"
)

func (h *FriendHandler) GetFriends(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	//1. always get the claims first authentication from middleware
	claims, ok := ctx.Value(middleware.UserContextKey).(*utils.JWTClaims)
	if !ok || claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	//2. check the user name from the details is that exist or not
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusUnauthorized)
		return
	}

	friends, err := h.friendService.GetFriends(ctx, userID)
	if err != nil {
		http.Error(w, "Failed to retrive Friends", http.StatusInternalServerError)
	}

	if friends == nil {
		friends = []dto.FriendResponse{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(friends)

}
