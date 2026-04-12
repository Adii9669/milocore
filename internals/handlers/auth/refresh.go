package auth

import (
	"chat-server/internals/middleware"
	"chat-server/internals/services"
	"chat-server/internals/utils"
	"net/http"
	"os"
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

		// 2. Get userID — refresh token is opaque so we need userID from
		//    the expired access token (still parseable even if expired)
		userID, err := middleware.GetUserIDFromContext(ctx)
		if err != nil {
			writeJSONError(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// 3. Issue new access token
		accessToken, err := authService.Refresh(ctx, userID, cookie.Value)
		if err != nil {
			writeJSONError(w, err.Error(), http.StatusUnauthorized)
			return
		}

		// 4. Set new access token cookie only
		isProduction := os.Getenv("APP_ENV") == "production"
		http.SetCookie(w, &http.Cookie{
			Name:     "access_token",
			Value:    accessToken,
			Path:     "/",
			MaxAge:   int(utils.AccessTokenExpiry.Seconds()),
			HttpOnly: true,
			Secure:   isProduction,
			SameSite: http.SameSiteLaxMode,
		})

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		utils.PrettyJSON(w, map[string]any{"message": "Token refreshed"})
	}
}
