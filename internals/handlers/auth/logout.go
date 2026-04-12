package auth

import (
	"chat-server/internals/middleware"
	"chat-server/internals/services"
	"chat-server/internals/utils"
	"net/http"
)

// func LogoutHandler(w http.ResponseWriter, r *http.Request) {
//
// 	//cretae a cookies which expires in the past
// 	expiredCookie := http.Cookie{
// 		Name:    "token",
// 		Value:   "",
// 		Expires: time.Now().Add(-1 * time.Hour),
// 		Path:    "/",
// 	}
//
// 	//set the cookies now
// 	http.SetCookie(w, &expiredCookie)
//
// 	// send the response
// 	w.WriteHeader(http.StatusOK)
// 	w.Write([]byte("Logged OUT"))
// }

func LogoutHandler(authService *services.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID, err := middleware.GetUserIDFromContext(ctx)
		if err != nil {
			writeJSONError(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		authService.Logout(ctx, userID)
		utils.ClearAuthCookies(w)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		utils.PrettyJSON(w, map[string]any{"message": "Logged out successfully"})
	}
}
