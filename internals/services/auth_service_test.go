package services

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"chat-server/internals/config"
	"chat-server/internals/db/models"
	"chat-server/internals/utils"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// 🔥 IMPORTANT: SET ENV FOR TESTS
func init() {
	config.Cfg.Secret.TOKEN = "test-secret"
	os.Setenv("TOKEN_SECRET", "test-secret")
}

//
// ================= MOCK USER REPO =================
//

type MockUserRepo struct {
	FindByEmailFn             func(ctx context.Context, email string) (*models.User, error)
	FindBYNameFn              func(ctx context.Context, name string) (*models.User, error)
	FindByIDFn                func(ctx context.Context, id uuid.UUID) (*models.User, error)
	CreateFn                  func(ctx context.Context, user *models.User) error
	ExistByIDFn               func(ctx context.Context, id uuid.UUID) (bool, error)
	ExistsByEmailOrUsernameFn func(ctx context.Context, email, username string) (bool, error)
	FindAllFn                 func(ctx context.Context) ([]models.User, error)
	FindByIDsFn               func(ctx context.Context, ids []uuid.UUID) ([]models.User, error)
	UpdateOTPFn               func(ctx context.Context, id uuid.UUID, hash, salt string, exp time.Time) error
	MarkVerifiedFn            func(ctx context.Context, id uuid.UUID) error
}

func (m *MockUserRepo) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	if m.FindByEmailFn != nil {
		return m.FindByEmailFn(ctx, email)
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *MockUserRepo) FindBYName(ctx context.Context, name string) (*models.User, error) {
	if m.FindBYNameFn != nil {
		return m.FindBYNameFn(ctx, name)
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *MockUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	if m.FindByIDFn != nil {
		return m.FindByIDFn(ctx, id)
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *MockUserRepo) Create(ctx context.Context, user *models.User) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, user)
	}
	return nil
}

func (m *MockUserRepo) ExistByID(ctx context.Context, id uuid.UUID) (bool, error) {
	if m.ExistByIDFn != nil {
		return m.ExistByIDFn(ctx, id)
	}
	return true, nil
}

func (m *MockUserRepo) ExistsByEmailOrUsername(ctx context.Context, email, username string) (bool, error) {
	if m.ExistsByEmailOrUsernameFn != nil {
		return m.ExistsByEmailOrUsernameFn(ctx, email, username)
	}
	return false, nil
}

func (m *MockUserRepo) FindAll(ctx context.Context) ([]models.User, error) {
	if m.FindAllFn != nil {
		return m.FindAllFn(ctx)
	}
	return []models.User{}, nil
}

func (m *MockUserRepo) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]models.User, error) {
	if m.FindByIDsFn != nil {
		return m.FindByIDsFn(ctx, ids)
	}
	return []models.User{}, nil
}

func (m *MockUserRepo) UpdateOTP(ctx context.Context, id uuid.UUID, hash, salt string, exp time.Time) error {
	if m.UpdateOTPFn != nil {
		return m.UpdateOTPFn(ctx, id, hash, salt, exp)
	}
	return nil
}

func (m *MockUserRepo) MarkVerifiedAndClearOTP(ctx context.Context, id uuid.UUID) error {
	if m.MarkVerifiedFn != nil {
		return m.MarkVerifiedFn(ctx, id)
	}
	return nil
}

//
// ================= MOCK SESSION REPO =================
//

type MockSessionRepo struct {
	CreateFn       func(ctx context.Context, s *models.Session) error
	FindByUserIDFn func(ctx context.Context, id uuid.UUID) (*models.Session, error)
	DeleteFn       func(ctx context.Context, id uuid.UUID) error
}

func (m *MockSessionRepo) Create(ctx context.Context, s *models.Session) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, s)
	}
	return nil
}

func (m *MockSessionRepo) FindByUserID(ctx context.Context, id uuid.UUID) (*models.Session, error) {
	if m.FindByUserIDFn != nil {
		return m.FindByUserIDFn(ctx, id)
	}
	return nil, errors.New("not found")
}

func (m *MockSessionRepo) DeleteByUserID(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, id)
	}
	return nil
}

func (m *MockSessionRepo) DeleteExpired(ctx context.Context) error {
	return nil
}

//
// ================= LOGIN TESTS =================
//

func TestLogin_Success(t *testing.T) {
	ctx := context.Background()

	password := "password123"
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	hashStr := string(hashed)

	mockUserRepo := &MockUserRepo{
		FindByEmailFn: func(ctx context.Context, email string) (*models.User, error) {
			return &models.User{
				ID:           uuid.New(),
				Email:        email,
				Name:         "milo",
				PasswordHash: &hashStr,
				Verified:     true,
			}, nil
		},
	}

	service := NewAuthService(mockUserRepo, &MockSessionRepo{}, nil)

	user, access, refresh, err := service.Login(ctx, "test@example.com", password)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.NotEmpty(t, access)
	assert.NotEmpty(t, refresh)
}

func TestLogin_InvalidPassword(t *testing.T) {
	ctx := context.Background()

	hashed, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.DefaultCost)
	hashStr := string(hashed)

	mockUserRepo := &MockUserRepo{
		FindByEmailFn: func(ctx context.Context, email string) (*models.User, error) {
			return &models.User{
				ID:           uuid.New(),
				PasswordHash: &hashStr,
				Verified:     true,
			}, nil
		},
	}

	service := NewAuthService(mockUserRepo, nil, nil)

	user, access, refresh, err := service.Login(ctx, "test@example.com", "wrong")

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Empty(t, access)
	assert.Empty(t, refresh)
}

//
// ================= VERIFY OTP TESTS =================
//

func TestVerifyOTP_Success(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	salt := "abc"
	otp := "123456"
	hash := utils.HashOTP(otp, salt)

	exp := time.Now().Add(5 * time.Minute)

	mockUserRepo := &MockUserRepo{
		FindByEmailFn: func(ctx context.Context, email string) (*models.User, error) {
			return &models.User{
				ID:            userID,
				Email:         email,
				Name:          "milo",
				VerifyOTPHash: hash,
				VerifySalt:    salt,
				OTPExpiresAt:  &exp,
				Verified:      false,
			}, nil
		},
	}

	service := NewAuthService(mockUserRepo, &MockSessionRepo{}, nil)

	user, access, refresh, err := service.VerifyOTP(ctx, "test@example.com", otp)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.True(t, user.Verified)
	assert.NotEmpty(t, access)
	assert.NotEmpty(t, refresh)
}

func TestVerifyOTP_InvalidOTP(t *testing.T) {
	ctx := context.Background()

	mockUserRepo := &MockUserRepo{
		FindByEmailFn: func(ctx context.Context, email string) (*models.User, error) {
			return &models.User{
				ID:            uuid.New(),
				VerifyOTPHash: "wrong",
				VerifySalt:    "abc",
				Verified:      false,
			}, nil
		},
	}

	service := NewAuthService(mockUserRepo, nil, nil)

	_, _, _, err := service.VerifyOTP(ctx, "test@example.com", "123456")

	assert.Error(t, err)
}

func TestVerifyOTP_Expired(t *testing.T) {
	ctx := context.Background()

	exp := time.Now().Add(-1 * time.Minute)

	mockUserRepo := &MockUserRepo{
		FindByEmailFn: func(ctx context.Context, email string) (*models.User, error) {
			return &models.User{
				ID:           uuid.New(),
				OTPExpiresAt: &exp,
				Verified:     false,
			}, nil
		},
	}

	service := NewAuthService(mockUserRepo, nil, nil)

	_, _, _, err := service.VerifyOTP(ctx, "test@example.com", "123456")

	assert.Error(t, err)
}
