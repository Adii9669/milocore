package auth

import (
	"chat-server/internals/middleware"
	"chat-server/internals/services"
	"chat-server/internals/utils"
	"net/http"
)

func LogoutHandler(authService *services.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID, err := middleware.GetUserIDFromContext(ctx)
		if err != nil {
			writeJSONError(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		if err := authService.Logout(ctx, userID, middleware.GetSessionIDFromContext(ctx)); err != nil {
			utils.ClearAuthCookies(w)
			writeJSONError(w, "Could not revoke this session", http.StatusServiceUnavailable)
			return
		}
		utils.ClearAuthCookies(w)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		utils.PrettyJSON(w, map[string]any{"message": "Logged out successfully"})
	}
}
