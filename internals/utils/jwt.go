package utils

import (
	"chat-server/internals/config"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims defines the structure of the data inside the token.
type JWTClaims struct {
	jwt.RegisteredClaims
	Username string `json:"name"`
	UserID   string `json:"userId"`
}

const (
	AccessTokenExpiry  = 15 * time.Minute
	RefreshTokenExpiry = 7 * 24 * time.Hour
)

// for getting the jwtkey
func jwtkey() ([]byte, error) {
	key := []byte(config.Cfg.Secret.TOKEN)
	if len(key) == 0 {
		return nil, fmt.Errorf("TOKEN secret is not set")
	}
	return key, nil
}

// GenerateAccessToken — short lived, 15 minutes
func GenerateAccessToken(userId uuid.UUID, username string) (string, error) {
	key, err := jwtkey()
	if err != nil {
		return "", err
	}

	claims := JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(AccessTokenExpiry)),
		},
		UserID:   userId.String(),
		Username: username,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(key)
}

// GenerateRefreshToken — random 32 byte hex string, NOT a JWT
// We store its hash in DB, send the raw value in the cookie
func GenerateRefreshToken() (raw string, hash string, err error) {

	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}
	raw = hex.EncodeToString(b)
	hash = HashOTP(raw, "refresh")
	return raw, hash, nil

}

// ValidateToken parses and decrypts a JWE token string.
// It returns the user claims if the token is valid, otherwise it returns an error.
func ValidateToken(tokenString string) (*JWTClaims, error) {

	//get the key
	key, err := jwtkey()
	if err != nil {
		return nil, err
	}

	// Parse the token with our custom claims struct.
	// The key function tells the parser how to get the secret key for verification.
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (any, error) {
		// Don't forget to validate the signing algorithm is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return key, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}
	// Check if the token is valid and extract the claims.
	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// kept for any existing callers — internally uses access token expiry now
func GenerateToken(userID uuid.UUID, name string) (string, error) {
	return GenerateAccessToken(userID, name)
}
