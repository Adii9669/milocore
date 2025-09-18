package auth

import (
	"encoding/json"
	"net/http"

	//internals
	"chat-server/internals/db"
	"chat-server/internals/utils"
	"chat-server/middleware"
)

func MeHandler(w http.ResponseWriter, r *http.Request) {

	//1.Get the user from context
	claims, ok := r.Context().Value(middleware.UserContextKey).(*utils.JWTClaims)
	if !ok {
		http.Error(w, "Could not retrive the data from the Context ", http.StatusInternalServerError)
		return
	}

	//2.Check the user you go from the key is in database or not
	user, err := db.FindUserByID(claims.Subject)
	if err != nil {
		http.Error(w, "Database Error", http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "Can't Find the USer", http.StatusNotFound)
		return
	}

	//3.Response it with user information
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"id":    user.ID,
		"name":  user.Name,
		"email": user.Email,
	})
}
