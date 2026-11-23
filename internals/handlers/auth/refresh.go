package auth

import (
	"chat-server/internals/services"
	"chat-server/internals/utils"
	"net/http"
)

// internals/handlers/auth/refresh.go
func RefreshHandler(authService *services.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// 1. Get refresh token from cookie
		cookie, err := r.Cookie("refresh_token")
		if err != nil {
			writeJSONError(w, "Missing refresh token", http.StatusUnauthorized)
			return
		}

		// 2. Resolve the session from the refresh token itself. The access
		//    token may already be expired/removed, so it cannot be used here.
		accessToken, err := authService.Refresh(ctx, cookie.Value)
		if err != nil {
			writeJSONError(w, "Invalid refresh token", http.StatusUnauthorized)
			return
		}

		// 3. Set new access token cookie only. Reuse the shared helper so the
		//    cookie keeps the right SameSite/Secure flags in production.
		utils.SetAccessCookie(w, accessToken)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		utils.PrettyJSON(w, map[string]any{"message": "Token refreshed"})
	}
}
