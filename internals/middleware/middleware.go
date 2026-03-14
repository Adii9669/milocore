package middleware

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"

	//project api
	"chat-server/internals/utils"

	"github.com/google/uuid"
)

// Define your context key.
// Unexported context key type (prevents collisions)
type contextKey struct{}

var userIDKey = contextKey{}

// AuthMiddleware uses the standard signature for gorilla/mux: func(http.Handler) http.Handler
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		tokenString := ""
		cookie, err := r.Cookie("token")
		if err == nil {
			tokenString = cookie.Value
		} else {
			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				tokenString = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		if tokenString == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		claims, err := utils.ValidateToken(tokenString)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, userID)

		refreshedToken, err := utils.GenerateToken(userID, claims.Username)
		if err == nil {
			isProduction := os.Getenv("APP_ENV") == "production"
			sameSite := http.SameSiteLaxMode
			secure := false
			if isProduction {
				secure = true
				sameSite = http.SameSiteNoneMode
			}
			http.SetCookie(w, &http.Cookie{
				Name:     "token",
				Value:    refreshedToken,
				Path:     "/",
				MaxAge:   int(utils.TokenExpiryDuration.Seconds()),
				HttpOnly: true,
				Secure:   secure,
				SameSite: sameSite,
			})
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	userID, ok := ctx.Value(userIDKey).(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.New("userID not found in context")
	}
	return userID, nil
}
