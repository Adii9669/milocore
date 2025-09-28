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

		//1.check the cookies first for the user
		cookie, err := r.Cookie("token")
		if err != nil {
			log.Printf("MIDDLEWARE ERROR: Cookie not found - %v", err)
			http.Error(w, "Unauthorized: Missing Auth Cookies", http.StatusUnauthorized)
			return
		}

		//2.now when the cookies are present check the token we stored in it
		tokenString := cookie.Value
		claims, err := utils.ValidateToken(tokenString)
		if err != nil {
			http.Error(w, "Token Invalid or Expired", http.StatusUnauthorized)
			return
		}

		//3.Now the token is present the create the context with the claims
		//From the details of that token and the context key you created
		ctx := context.WithValue(r.Context(), UserContextKey, claims)

		//4.Pass the  modified request to the next handler.
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
