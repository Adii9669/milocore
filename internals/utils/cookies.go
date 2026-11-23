package utils

import (
	"net/http"
	"os"
	"time"
)

// cookieSettings keeps the access and refresh cookies consistent. In
// production the frontend and API are on different sites, so the refresh
// endpoint must re-issue the access cookie with SameSite=None; Secure.
func cookieSettings() (secure bool, sameSite http.SameSite) {
	isProduction := os.Getenv("APP_ENV") == "production"
	if isProduction {
		return true, http.SameSiteNoneMode
	}
	return false, http.SameSiteLaxMode
}

// SetAccessCookie sets the short-lived access token cookie. Used by login,
// OTP verification and token refresh.
func SetAccessCookie(w http.ResponseWriter, accessToken string) {
	secure, sameSite := cookieSettings()

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		MaxAge:   int(AccessTokenExpiry.Seconds()),
		Expires:  time.Now().Add(AccessTokenExpiry),
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	})
}

func SetAuthCookies(w http.ResponseWriter, accessToken, refreshToken string) {
	secure, sameSite := cookieSettings()

	SetAccessCookie(w, accessToken)

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/auth/refresh", // scoped — only sent to refresh endpoint
		MaxAge:   int(RefreshTokenExpiry.Seconds()),
		Expires:  time.Now().Add(RefreshTokenExpiry),
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	})
}

func ClearAuthCookies(w http.ResponseWriter) {
	secure, sameSite := cookieSettings()

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/auth/refresh",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	})
}
