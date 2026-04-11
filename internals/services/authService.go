package services

import (
	"chat-server/internals/db/models"
	"chat-server/internals/repository"
	"chat-server/internals/utils"
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	userRepo     repository.UserRepository
	emailService *EmailService
}

func NewAuthService(userRepo repository.UserRepository, emailService *EmailService) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		emailService: emailService,
	}
}

func (s *AuthService) Register(ctx context.Context, username, email, password string) error {

	// 1. check existing user
	existingUser, err := s.userRepo.FindByEmail(ctx, email)
	if err == nil {
		if existingUser.Verified {
			return fmt.Errorf("user already exists")
		}

		otp, err := utils.GenerateOtp()
		if err != nil {
			return err
		}

		//hasing the otp
		salt := uuid.NewString()
		otpHash := utils.HashOTP(otp, salt)
		expiresAt := time.Now().Add(5 * time.Minute)

		err = s.userRepo.UpdateOTP(ctx, existingUser.ID, otpHash, salt, expiresAt)
		if err != nil {
			return err
		}

		err = s.emailService.SendOTP(email, otp)
		if err != nil {
			log.Println("email resend failed:", err)
		}

		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err // real DB error
	}

	// 2. Check username uniqueness
	_, err = s.userRepo.FindBYName(ctx, username)
	if err == nil {
		return fmt.Errorf("username already taken")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	//2.Hash the Password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	hashedPasswordStr := string(passwordHash)

	//3.generate OTP
	otp, err := utils.GenerateOtp()
	if err != nil {
		return err
	}

	salt := uuid.NewString()
	otpHash := utils.HashOTP(otp, salt)

	expiresAt := time.Now().Add(5 * time.Minute)

	//4.Create the User
	newUser := models.User{
		Name:          username,
		Email:         email,
		PasswordHash:  &hashedPasswordStr,
		Verified:      false,
		VerifyOTPHash: otpHash,
		VerifySalt:    salt,
		OTPExpiresAt:  &expiresAt,
	}

	// 5.Save the User
	err = s.userRepo.Create(ctx, &newUser)
	if err != nil {
		return err
	}

	//6.async mail
	go func() {
		err := s.emailService.SendOTP(email, otp)
		if err != nil {
			log.Println("failed to send the otp:", err)
		}
	}()

	return nil
}

func (s *AuthService) VerifyOTP(ctx context.Context, email, otp string) (*models.User, error) {
	// 1. Find user
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	// 2. Already verified?
	if user.Verified {
		return nil, fmt.Errorf("account is already verified")
	}

	// 3. Check OTP expiry
	if user.OTPExpiresAt == nil || time.Now().After(*user.OTPExpiresAt) {
		return nil, fmt.Errorf("verification code has expired")
	}

	// 4. Hash the incoming OTP with the stored salt, then compare
	if user.VerifyOTPHash == "" || user.VerifySalt == "" {
		return nil, fmt.Errorf("no verification code found")
	}
	submittedHash := utils.HashOTP(otp, user.VerifySalt)
	if submittedHash != user.VerifyOTPHash {
		return nil, fmt.Errorf("invalid verification code")
	}

	// 5. Mark verified and clear OTP fields via repository
	err = s.userRepo.MarkVerifiedAndClearOTP(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	user.Verified = true
	return user, nil
}
