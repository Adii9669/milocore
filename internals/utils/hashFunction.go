package utils

import (
	"crypto/sha256"
	"encoding/hex"
)

func HashOTP(otp string, salt string) string {
	//combine otp + salt
	data := otp + salt

	hash := sha256.Sum256([]byte(data))

	return hex.EncodeToString(hash[:])
}
