package utils

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims defines the structure of the data inside the token.
type JWTClaims struct {
	jwt.RegisteredClaims
	Name string `json:"name"`
}

func GenerateToken(userId string, name string) (string, error) {
	//take the token key
	jwtkey := []byte(os.Getenv("TOKEN_KEY"))
	if len(jwtkey) == 0 {
		return "", fmt.Errorf("JWT_SECRET_KEY environment variable not set")
	}

	// Create the claims for the token.
	claims := JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			// 'Subject' is the standard claim for the user's unique identifier.
			Subject: userId,
			// 'ExpiresAt' is the standard claim for the token's expiration time.
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
		Name: name,
	}

	//create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	//sign in with the token
	signedToken, err := token.SignedString(jwtkey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil

}

// ValidateToken parses and decrypts a JWE token string.
// It returns the user claims if the token is valid, otherwise it returns an error.
func ValidateToken(tokenString string) (*JWTClaims, error) {

	//get the key
	jwtKey := []byte(os.Getenv("TOKEN_KEY"))
	if len(jwtKey) == 0 {
		return nil, fmt.Errorf("KEY env not set")
	}

	// Parse the token with our custom claims struct.
	// The key function tells the parser how to get the secret key for verification.
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (any, error) {
		// Don't forget to validate the signing algorithm is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtKey, nil
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
