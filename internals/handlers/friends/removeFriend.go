package friends

import (
	"chat-server/internals/middleware"
	"chat-server/internals/utils"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *FriendHandler) RemoveFriend(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()
	ctx := r.Context()

	// 1️⃣ Get authenticated user from context
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		http.Error(w, "User ID not found", http.StatusUnauthorized)
		return
	}
	fmt.Println("UserID", userID)

	//3. Extract FriendID using Gorilla Mux
	friendIDParam := chi.URLParam(r, "friendID")
	fmt.Println("FriendID from URL:", friendIDParam)
	friendID, err := uuid.Parse(friendIDParam)
	if err != nil {
		http.Error(w, "invalid friend id", http.StatusBadRequest)
		return
	}

	//4. Delete the Friend
	err = h.friendService.RemoveFriend(ctx, userID, friendID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	utils.PrettyJSON(w, "Friend Removed")
}
