package services

import (
	"chat-server/internals/utils"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

func (s *AuthService) ResendOTP(ctx context.Context, email string) error {

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return err
	}

	if user.Verified {
		return errors.New("user already verified")
	}

	otp, _ := utils.GenerateOtp()
	salt := uuid.NewString()
	otpHash := utils.HashOTP(otp, salt)

	expiresAt := time.Now().Add(5 * time.Minute)

	err = s.userRepo.UpdateOTP(ctx, user.ID, otpHash, salt, expiresAt)
	if err != nil {
		return err
	}

	return s.emailService.SendOTP(email, otp)
}
