package middleware

import (
	"context"
	"errors"
	"net/http"

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

		//1.Extract token from the cookies
		cookie, err := r.Cookie("token")
		if err != nil {
			http.Error(w, "Unauthorized ", http.StatusUnauthorized)
			return
		}

		//2.Validate the Token
		tokenString := cookie.Value
		claims, err := utils.ValidateToken(tokenString)
		if err != nil {
			http.Error(w, "Unauthorized ", http.StatusUnauthorized)
			return
		}

		// Parse UUID (normalize identity)
		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			http.Error(w, "Unauthorized ", http.StatusUnauthorized)
			return
		}

		//3.Now the token is present the create the context with the claims
		ctx := context.WithValue(r.Context(), userIDKey, userID)

		//4.Pass the  modified request to the next handler.
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
