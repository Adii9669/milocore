package friends

import (
	"chat-server/internals/utils"
	"chat-server/middleware"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

func (h *FriendHandler) SendFrndRequest(w http.ResponseWriter, r *http.Request) {

	//1.Check the body of the client (Auth)
	claims, ok := r.Context().Value(middleware.UserContextKey).(*utils.JWTClaims)
	if !ok || claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	//2.get the userID and stroing it converting the userid(string) to the claims uuid
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		http.Error(w, "Invalid user id token", http.StatusUnauthorized)
		return
	}

	//3.Decode the body with info required
	var body struct {
		FriendID uuid.UUID `json:"friendId"`
	}
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

	//validate targetUser
	//fist check if the user exist or not (to the user you are sending the request )
	targetUser, err := h.UserRepo.FindByID(body.FriendID)
	if err != nil {
		http.Error(w, "User NOT FOUND", http.StatusNotFound)
		return
	}

	if targetUser == nil {
		http.Error(w, "User Not Found", http.StatusNotFound)
		return
	}

	//Check for the reverse request for the user whom you sending the friend request
	if rel, err := h.FrndRepo.FindRelation(body.FriendID, userID); err == nil && rel != nil {
		if rel.Status == "pending" {

			//accecpt the pending request
			if err := h.FrndRepo.AcceptRequest(rel); err != nil {
				http.Error(w, "Failed To Accept FriendRequest", http.StatusInternalServerError)
				return
			}
			// And create the reverse accepted row (userID -> friend)
			if err := h.FrndRepo.CreateReverseAccepted(userID, body.FriendID); err != nil {
				http.Error(w, "Failed to finalize acceptance", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			utils.PrettyJSON(w, map[string]string{"message": "Friend request accepted (mutual)"})
			return
		}

		if rel.Status == "accepted" {
			http.Error(w, "Already Friend", http.StatusBadRequest)
			return
		}

	}

	// Now check if the caller already sent a request (user -> friend)
	if rel2, err := h.FrndRepo.FindRelation(userID, body.FriendID); err == nil && rel2 != nil {
		// if a pending request already exists
		if rel2.Status == "pending" {
			http.Error(w, "Friend request already sent", http.StatusBadRequest)
			return
		}
		// if relation already accepted
		if rel2.Status == "accepted" {
			http.Error(w, "Already friends", http.StatusBadRequest)
			return
		}
	}

	//  No relation exists -> create a PENDING request (user -> friend)
	if err := h.FrndRepo.SendRequest(userID, body.FriendID); err != nil {
		// if unique constraint violation happens, return 400 or appropriate message
		http.Error(w, "Failed to send friend request", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	utils.PrettyJSON(w, map[string]string{
		"message": "Friend request sent",
	})

}
