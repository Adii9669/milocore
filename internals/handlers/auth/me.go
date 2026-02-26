package auth

import (
	"net/http"

	//internals
	"chat-server/internals/middleware"
	"chat-server/internals/utils"
)

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	//2.Check the user you go from the key is in database or not
	user, err := h.UserRepo.FindByID(ctx, userID)
	if err != nil {
		http.Error(w, "Can't Find the USer", http.StatusNotFound)
		return
	}

	//3.Response it with user information
	w.Header().Set("Content-Type", "application/json")
	utils.PrettyJSON(w, map[string]any{
		"message": "User Details",
		"user": map[string]any{
			"id":     user.ID,
			"name":   user.Name,
			"email":  user.Email,
			"status": "verified",
		},
	})
}
