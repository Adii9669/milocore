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
		log.Printf("MIDDLEWARE: Successfully found cookie named '%s'", cookie.Name)

		tokenString := cookie.Value
		log.Println("Token middleware", tokenString)
		claims, err := utils.ValidateToken(tokenString)
		if err != nil {
			http.Error(w, "Token Invalid or Expired", http.StatusUnauthorized)
			return
		}
		log.Printf("MIDDLEWARE: Successfully validated token. Setting claims in context: %+v", claims)

		// Create a new context with the claims.
		ctx := context.WithValue(r.Context(), UserContextKey, claims)

		// This is the correct way to pass the modified request to the next handler.
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// package middleware
//
// import (
// 	"context"
//
// 	"fmt"
//
// 	"log"
//
// 	"net/http"
//
// 	"github.com/gorilla/mux"
//
// 	"chat-server/internals/utils"
// )
//
// type Middleware func(http.HandlerFunc) http.HandlerFunc
//
// // a custom constext for avoding collision
//
// type ContextKey string
//
// const UserContextKey = ContextKey("key")
//
// func AuthMiddleware() Middleware {
//
// 	//create a new middleware
//
// 	return func(next http.HandlerFunc) http.HandlerFunc {
//
// 		//Define the http handler func
//
// 		return func(w http.ResponseWriter, r *http.Request) {
//
// 			// get the cookie from the request
//
// 			cookie, err := r.Cookie("token")
//
// 			log.Println("Cookie =", cookie)
//
// 			if err != nil {
//
// 				log.Printf("Middleware Error: Cookie not found %v", err)
//
// 				http.Error(w, "Unauthorized: Missing Auth Cookies", http.StatusUnauthorized)
//
// 				return
//
// 			}
//
// 			if cookie == nil {
//
// 				log.Printf("The Cookies are empty %v", cookie)
//
// 				return
//
// 			}
//
// 			// take the token from the cookie
//
// 			tokenString := cookie.Value
//
// 			// log.Printf("THe COOKIE value is  %v", tokenString)
//
// 			//take this token validate it and store the claims which comes in return from the validation file
//
// 			claims, err := utils.ValidateToken(tokenString)
//
// 			fmt.Println("Middleware claims:------", claims)
//
// 			if err != nil {
//
// 				http.Error(w, "Token Invalid or Expired ", http.StatusUnauthorized)
//
// 				return
//
// 			}
//
// 			//Token is valid and store the user claims in the request context
//
// 			ctx := context.WithValue(r.Context(), UserContextKey, claims)
//
// 			next(w, r.WithContext(ctx))
//
// 		}
//
// 	}
//
// }
//
// func MuxAdapter(m Middleware) mux.MiddlewareFunc {
//
// 	return func(next http.Handler) http.Handler {
//
// 		return m(next.ServeHTTP)
//
// 	}
//
// }
