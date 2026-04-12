package auth

import (
	"chat-server/internals/services"
	"encoding/json"
	"net/http"
)

func ResendOTPHandler(authService *services.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Email string `json:"email"`
		}

		json.NewDecoder(r.Body).Decode(&req)

		err := authService.ResendOTP(r.Context(), req.Email)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
