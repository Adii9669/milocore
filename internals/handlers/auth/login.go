package auth

import (
	"encoding/json"
	"net/http"
	"strings"

	// internals
	"chat-server/internals/requests"
	"chat-server/internals/services"
	"chat-server/internals/utils"
)

// helper Function to write the json errors
func writeJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"message": message})
}

func LoginHandler(authService *services.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var req requests.Credentials
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, "Invalid request Body.", http.StatusBadRequest)
			return
		}

		req.Username = strings.TrimSpace(req.Username)
		if req.Username == "" || req.Password == "" {
			writeJSONError(w, "Empty Fields.", http.StatusBadRequest)
			return
		}

		user, accessToken, refreshToken, err := authService.Login(ctx, req.Username, req.Password)
		if err != nil {
			switch err.Error() {
			case "invalid credentials":
				writeJSONError(w, err.Error(), http.StatusUnauthorized)
			case "account not verified":
				writeJSONError(w, err.Error(), http.StatusForbidden)
			default:
				writeJSONError(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}
		// 3. Set both cookies — replaces the old single token block
		utils.SetAuthCookies(w, accessToken, refreshToken)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		utils.PrettyJSON(w, map[string]any{
			"message": "Logged In Successfully",
			"user": map[string]any{
				"id":     user.ID,
				"name":   user.Name,
				"email":  user.Email,
				"status": "verified",
			},
		})
	}
}
