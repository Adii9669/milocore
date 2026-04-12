package utils

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"
)

// Gnereate the TOken
func GenerateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateOtp
func GenerateOtp() (string, error) {
	const otpLength = 6

	otp := make([]byte, otpLength)

	for i := 0; i < otpLength; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		otp[i] = byte('0' + n.Int64())
	}

	return string(otp), nil
}
