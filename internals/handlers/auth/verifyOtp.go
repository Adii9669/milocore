package auth

import (
	"chat-server/internals/requests"
	"chat-server/internals/services"
	"chat-server/internals/utils"
	"encoding/json"
	"net/http"
)

func VerifyOtpHandler(authService *services.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// 1. Decode request
		var req requests.VerifyOtpRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// 2. Delegate all logic to the service
		user, accessToken, refreshToken, err := authService.VerifyOTP(ctx, req.Email, req.OTP)
		if err != nil {
			status := http.StatusInternalServerError
			msg := err.Error()
			switch msg {
			case "user not found":
				status = http.StatusNotFound
			case "account is already verified",
				"invalid verification code",
				"no verification code found":
				status = http.StatusBadRequest
			case "verification code has expired":
				status = http.StatusUnprocessableEntity
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(status)
			json.NewEncoder(w).Encode(map[string]string{"message": msg})
			return
		}
		// 3. Set both cookies — replaces the old single token block
		utils.SetAuthCookies(w, accessToken, refreshToken)

		// 4. Respond
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		utils.PrettyJSON(w, map[string]any{
			"message":  "Account verified successfully.",
			"verified": user.Verified,
		})
	}
}
