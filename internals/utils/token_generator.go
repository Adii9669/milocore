package utils

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"
)

func GenerateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// func GenerateOTP() string {
// 	max := big.NewInt(10000) // 0–9999
// 	n, _ := rand.Int(rand.Reader, max)
// 	return fmt.Sprintf("%04d", n.Int64())
// }

func GenerateOtp() (string, error) {
	const otpLength = 6
	var otp string
	for i := 0; i < otpLength; i++ {
		digit, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			// Handle the error, as rand.Int can return one
			return "", err
		}

		otp += digit.String()
	}
	return otp, nil
}
