package services

import (
	"chat-server/internals/db/models"
	"chat-server/internals/repository"
	"chat-server/internals/utils"
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	userRepo     repository.UserRepository
	sessionRepo  repository.SessionRepository
	emailService *EmailService
}

func NewAuthService(
	userRepo repository.UserRepository,
	sessionRepo repository.SessionRepository,
	emailService *EmailService) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		emailService: emailService,
		sessionRepo:  sessionRepo,
	}
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrNotVerified        = errors.New("account not verified")
)

var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%^&*+\\/?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

func isEmail(e string) bool {
	return emailRegex.MatchString(e)
}

// Add this helper — called by both Login and VerifyOTP
func (s *AuthService) createSession(
	ctx context.Context,
	user *models.User,
) (accessToken, refreshToken string, err error) {
	// 1. Generate access token
	accessToken, err = utils.GenerateAccessToken(user.ID, user.Name)
	if err != nil {
		return "", "", err
	}

	//2.Generate the token
	rawRefresh, refreshHash, err := utils.GenerateRefreshToken()
	if err != nil {
		return "", "", err
	}

	//3.Store the hash in the DB
	expiresAt := time.Now().Add(utils.RefreshTokenExpiry)
	session := &models.Session{
		UserID:           user.ID,
		RefreshTokenHash: refreshHash,
		ExpiresAt:        expiresAt,
	}
	if err = s.sessionRepo.Create(ctx, session); err != nil {
		return "", "", fmt.Errorf("failed to create session: %w", err)
	}
	return accessToken, rawRefresh, nil
}

// Refresh — validates refresh token and issues new access token
func (s *AuthService) Refresh(ctx context.Context, userID uuid.UUID, rawRefreshToken string) (string, error) {
	session, err := s.sessionRepo.FindByUserID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("session not found")
	}

	// hash the incoming token and compare
	incomingHash := utils.HashOTP(rawRefreshToken, "refresh")
	if incomingHash != session.RefreshTokenHash {
		return "", fmt.Errorf("invalid refresh token")
	}

	// fetch user for name claim
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("user not found")
	}

	return utils.GenerateAccessToken(user.ID, user.Name)
}

func (s *AuthService) Login(ctx context.Context, username, password string) (*models.User, string, string, error) {

	var user *models.User
	var err error

	//1.Find the user
	if isEmail(username) {
		user, err = s.userRepo.FindByEmail(ctx, username)
	} else {
		user, err = s.userRepo.FindBYName(ctx, username)
	}
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, "", "", ErrInvalidCredentials
		}
		return nil, "", "", err
	}

	//2.check password
	if err := bcrypt.CompareHashAndPassword(
		[]byte(*user.PasswordHash),
		[]byte(password)); err != nil {
		return nil, "", "", ErrInvalidCredentials
	}

	//3.check verification is verified or not
	if !user.Verified {
		return nil, "", "", ErrNotVerified
	}

	// userID, err := uuid.Parse(user.ID.String())
	// if err != nil {
	// 	log.Println("parsing failed for userID")
	// 	return nil, "", errors.New("parsing failed")
	// }

	//4.Generate the Token
	// tokenString, err := utils.GenerateToken(userID, user.Email)
	// if err != nil {
	// 	return nil, "", err
	// }

	// return user, tokenString, nil
	accessToken, refreshToken, err := s.createSession(ctx, user)
	if err != nil {
		return nil, "", "", err
	}

	return user, accessToken, refreshToken, nil
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

// VerifyOTP
func (s *AuthService) VerifyOTP(
	ctx context.Context,
	email, otp string) (*models.User, string, string, error) {

	// 1. Find user
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", "", fmt.Errorf("user not found")
		}
		return nil, "", "", fmt.Errorf("database error: %w", err)
	}

	// 2. Already verified?
	if user.Verified {
		return nil, "", "", ErrNotVerified
	}

	// 3. Check OTP expiry
	if user.OTPExpiresAt == nil || time.Now().After(*user.OTPExpiresAt) {
		return nil, "", "", fmt.Errorf("verification code has expired")
	}

	// 4. Hash the incoming OTP with the stored salt, then compare
	if user.VerifyOTPHash == "" || user.VerifySalt == "" {
		return nil, "", "", fmt.Errorf("no verification code found")
	}
	submittedHash := utils.HashOTP(otp, user.VerifySalt)
	if submittedHash != user.VerifyOTPHash {
		return nil, "", "", fmt.Errorf("invalid verification code")
	}

	// 5. Mark verified and clear OTP fields via repository
	err = s.userRepo.MarkVerifiedAndClearOTP(ctx, user.ID)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to update user: %w", err)
	}

	user.Verified = true

	// 6. Create session — NEW
	accessToken, refreshToken, err := s.createSession(ctx, user)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to create session: %w", err)
	}

	return user, accessToken, refreshToken, nil
}

// Logout — deletes session from DB
func (s *AuthService) Logout(ctx context.Context, userID uuid.UUID) error {
	return s.sessionRepo.DeleteByUserID(ctx, userID)
}
