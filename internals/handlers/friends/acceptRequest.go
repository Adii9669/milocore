package friends

import (
	"chat-server/internals/requests"
	"chat-server/internals/utils"
	"chat-server/middleware"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

func (h *FriendHandler) AcceptRequest(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	//1.Auth check the client
	claims, ok := ctx.Value(middleware.UserContextKey).(*utils.JWTClaims)
	if !ok || claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		http.Error(w, "Invalid user id token", http.StatusUnauthorized)
		return
	}

	//2.Decode the Body and validate it
	var body requests.FriendRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid Request", http.StatusBadRequest)
		return
	}

	friendID := body.FriendID

	if friendID == uuid.Nil {
		http.Error(w, "FriendRequest ID required", http.StatusBadRequest)
		return
	}

	if friendID == userID {
		http.Error(w, "cannot accept your own request", http.StatusBadRequest)
		return
	}

	//Here we call our friendService to do futher task on the request
	err = h.friendService.AcceptFriendRequest(ctx, userID, friendID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	utils.PrettyJSON(w, map[string]string{
		"message": "Friend request accepted",
	})
}
