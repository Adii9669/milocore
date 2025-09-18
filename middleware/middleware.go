package middleware

import (
	"context"
	"log"
	"net/http"

	//project api
	"chat-server/internals/utils"
)

// Define your context key.
type contextKey string

const UserContextKey = contextKey("user_claims")

// AuthMiddleware uses the standard signature for gorilla/mux: func(http.Handler) http.Handler
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("token")

		if err != nil {
			log.Printf("MIDDLEWARE ERROR: Cookie not found - %v", err)
			http.Error(w, "Unauthorized: Missing Auth Cookies", http.StatusUnauthorized)
			return
		}
		// log.Printf("MIDDLEWARE: Successfully found cookie named '%s'", cookie.Name)

		tokenString := cookie.Value
		// log.Println("Token middleware", tokenString)
		claims, err := utils.ValidateToken(tokenString)
		if err != nil {
			http.Error(w, "Token Invalid or Expired", http.StatusUnauthorized)
			return
		}
		// log.Printf("MIDDLEWARE: Successfully validated token. Setting claims in context: %+v", claims)

		// Create a new context with the claims.
		ctx := context.WithValue(r.Context(), UserContextKey, claims)

		// This is the correct way to pass the modified request to the next handler.
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
