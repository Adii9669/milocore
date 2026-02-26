package friends

import (
	"chat-server/internals/middleware"
	"chat-server/internals/requests"
	"chat-server/internals/utils"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

func (h *FriendHandler) SendFrndRequest(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		http.Error(w, "Invalid user id token", http.StatusUnauthorized)
		return
	}

	//3.Decode the body with info required
	var body requests.FriendRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request Body", http.StatusBadRequest)
		return
	}

	//Body Validation
	// checking if friendID is nil
	if body.FriendID == uuid.Nil {
		http.Error(w, "Friend ID is required ", http.StatusBadRequest)
		return
	}

	// checking to check if not sending request to self
	if body.FriendID == userID {
		http.Error(w, "Can NOt Send the Friend Request to yourself", http.StatusBadRequest)
		return
	}

	err = h.friendService.SendFriendRequest(ctx, userID, body.FriendID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	utils.PrettyJSON(w, map[string]string{
		"response": "Friend request sent",
	})

}
