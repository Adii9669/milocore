package utils

import (
	"net/http"
	"os"
)

func SetAuthCookies(w http.ResponseWriter, accessToken, refreshToken string) {
	isProduction := os.Getenv("APP_ENV") == "production"
	secure := isProduction
	sameSite := http.SameSiteLaxMode
	if isProduction {
		sameSite = http.SameSiteNoneMode
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		MaxAge:   int(AccessTokenExpiry.Seconds()),
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/auth/refresh", // scoped — only sent to refresh endpoint
		MaxAge:   int(RefreshTokenExpiry.Seconds()),
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	})
}

func ClearAuthCookies(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: "access_token", MaxAge: -1, Path: "/"})
	http.SetCookie(w, &http.Cookie{Name: "refresh_token", MaxAge: -1, Path: "/auth/refresh"})
}
