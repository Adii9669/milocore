package utils

import (
	"testing"

	"github.com/go-jose/go-jose/v4/testutils/require"
	"github.com/stretchr/testify/assert"
)

func TestHashOTP_Deterministic(t *testing.T) {
	hash1 := HashOTP("123456", "mysalt")
	hash2 := HashOTP("123456", "mysalt")
	assert.Equal(t, hash1, hash2, "HashtOTP should be deterministic")
}

func TestHashOTP_DifferentSaltDiffHash(t *testing.T) {
	hash1 := HashOTP("123456", "salt-one")
	hash2 := HashOTP("123456", "salt-two")
	assert.NotEqual(t, hash1, hash2, "HashtOTP should be deterministic")
}

func TestHashOTP_DifferentOTPDiffHash(t *testing.T) {
	hash1 := HashOTP("123456", "samesalt")
	hash2 := HashOTP("654321", "samesalt")
	assert.NotEqual(t, hash1, hash2, "Different OTPs must produce different hashes")
}

func TestHashOTP_ReturnsHexString(t *testing.T) {
	hash := HashOTP("123456", "mysalt")
	// SHA256 hex output is always 64 characters
	assert.Len(t, hash, 64, "SHA256 hex output should be 64 chars")
}

func TestHashOTP_EmptyOTP(t *testing.T) {
	// Should not panic, just hash empty string + salt
	hash := HashOTP("", "mysalt")
	assert.Len(t, hash, 64, "Should still return valid hash for empty OTP")
}

func TestHashOTP_EmptySalt(t *testing.T) {
	hash := HashOTP("123456", "")
	assert.Len(t, hash, 64, "Should still return valid hash for empty salt")
}

func TestHashOTP_SaltActuallyMatters(t *testing.T) {
	// This is the CRITICAL security test
	// Proves that two users with same OTP get different hashes
	// because of different salts (UUID per user)
	userOneSalt := "a1b2c3d4-uuid-user-one"
	userTwoSalt := "z9y8x7w6-uuid-user-two"

	hash1 := HashOTP("123456", userOneSalt)
	hash2 := HashOTP("123456", userTwoSalt)

	assert.NotEqual(t, hash1, hash2,
		"Same OTP with different salts must never produce same hash — rainbow table protection")
}

// ─────────────────────────────────────────
// GenerateOtp Tests
// ─────────────────────────────────────────

func TestGenerateOTP_length(t *testing.T) {
	otp, err := GenerateOtp()
	require.NoError(t, err, "generateOTP should not return error")
	assert.Len(t, otp, 6, "OTP must be exactly 6 digits")
}

func TestGenerateOtp_OnlyDigits(t *testing.T) {
	otp, err := GenerateOtp()
	require.NoError(t, err)

	for i, ch := range otp {
		assert.True(t, ch >= '0' && ch <= '9',
			"Character at index %d is not a digit: %c", i, ch)
	}
}

func TestGenerateOtp_Uniqueness(t *testing.T) {
	// Generate 100 OTPs and check they're not all identical
	// (they CAN collide by chance but 100 identical ones = broken)
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		otp, err := GenerateOtp()
		require.NoError(t, err)
		seen[otp] = true
	}
	assert.Greater(t, len(seen), 1, "100 generated OTPs should not all be identical")
}

func TestGenerateOtp_NoError(t *testing.T) {
	// Run 10 times to confirm crypto/rand never errors in normal conditions
	for i := 0; i < 10; i++ {
		_, err := GenerateOtp()
		require.NoError(t, err)
	}
}
