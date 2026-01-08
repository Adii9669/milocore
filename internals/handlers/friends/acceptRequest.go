package friends

import (
	"chat-server/internals/requests"
	"chat-server/internals/utils"
	"chat-server/middleware"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (h *FriendHandler) AcceptRequest(w http.ResponseWriter, r *http.Request) {

	//1.Auth check the client
	claims, ok := r.Context().Value(middleware.UserContextKey).(*utils.JWTClaims)
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

	if body.FriendID == uuid.Nil {
		http.Error(w, "FriendRequest ID required", http.StatusBadRequest)
		return
	}

	if body.FriendID == userID {
		http.Error(w, "cannot accept your own request", http.StatusBadRequest)
		return
	}

	//Ceck the status
	rel, err := h.FrndRepo.FindRelation(body.FriendID, userID)
	if err != nil {
		// If no row found: return 404
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Friend request not found", http.StatusNotFound)
			return
		}
		// DB error
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if rel == nil {
		http.Error(w, "Database Error", http.StatusNotFound)
		return
	}

	if rel.Status != "pending" {
		if rel.Status == "accepted" {
			http.Error(w, "Already friends", http.StatusBadRequest)
			return
		}
		http.Error(w, "Invalid request status", http.StatusBadRequest)
		return
	}

	// 5) Accept and create reverse accepted row ATOMICALLY
	if err := h.FrndRepo.AcceptAndCreateReverse(rel, userID); err != nil {
		// If unique constraint on reverse exists, the repo may return an error.
		// Surface as 500 for now; you can map DB-specific errors to friendly codes.
		http.Error(w, "Failed to accept friend request", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	utils.PrettyJSON(w, map[string]string{"message": "Friend request accepted"})
}
