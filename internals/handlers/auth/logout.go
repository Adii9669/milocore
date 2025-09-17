package auth

import (
	"net/http"
	"time"
)

func LogoutHandler(w http.ResponseWriter, r *http.Request) {

	//cretae a cookies which expires in the past
	expiredCookie := http.Cookie{
		Name:    "token",
		Value:   "",
		Expires: time.Now().Add(-1 * time.Hour),
		Path:    "/",
	}

	//set the cookies now
	http.SetCookie(w, &expiredCookie)

	// send the response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Logged OUT"))
}
