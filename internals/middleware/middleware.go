package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	//project api
	"chat-server/internals/utils"

	"github.com/google/uuid"
)

// Define your context key.
// Unexported context key type (prevents collisions)
type contextKey struct{}

var userIDKey = contextKey{}

type sessionContextKey struct{}

// AuthMiddleware uses the standard signature for gorilla/mux: func(http.Handler) http.Handler
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		tokenString := ""

		// check access_token cookie first
		cookie, err := r.Cookie("access_token")
		if err == nil {
			tokenString = cookie.Value
		} else {
			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				tokenString = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		if tokenString == "" {
			// The access cookie is gone (its 15-minute MaxAge elapsed). Tell
			// the client to use its refresh cookie instead of logging out.
			w.Header().Set("X-Token-Expired", "true")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		claims, err := utils.ValidateToken(tokenString)
		if err != nil {
			w.Header().Set("X-Token-Expired", "true")
			http.Error(w, "Token expired", http.StatusUnauthorized)
			return
		}

		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			w.Header().Set("X-Token-Expired", "true")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), userIDKey, userID)
		if sessionID, err := uuid.Parse(claims.ID); err == nil {
			ctx = context.WithValue(ctx, sessionContextKey{}, sessionID)
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

func GetSessionIDFromContext(ctx context.Context) uuid.UUID {
	id, _ := ctx.Value(sessionContextKey{}).(uuid.UUID)
	return id
}
