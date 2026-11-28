package push

import (
	"chat-server/internals/db/models"
	"chat-server/internals/middleware"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type deviceStore interface {
	register(context.Context, models.PushDevice) error
	remove(context.Context, uuid.UUID, uuid.UUID) error
}

type Store struct{ db *gorm.DB }

func NewStore(db *gorm.DB) *Store { return &Store{db: db} }

func (s *Store) register(ctx context.Context, device models.PushDevice) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// A token may be reassigned on reinstall or account change.
		if err := tx.Where("token = ? AND installation_id <> ?", device.Token, device.InstallationID).
			Delete(&models.PushDevice{}).Error; err != nil {
			return err
		}
		return tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "installation_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"user_id", "token", "updated_at"}),
		}).Create(&device).Error
	})
}

func (s *Store) remove(ctx context.Context, userID, installationID uuid.UUID) error {
	return s.db.WithContext(ctx).Where("user_id = ? AND installation_id = ?", userID, installationID).
		Delete(&models.PushDevice{}).Error
}

func (s *Store) forUsers(ctx context.Context, users []uuid.UUID) ([]models.PushDevice, error) {
	var devices []models.PushDevice
	err := s.db.WithContext(ctx).Where("user_id IN ?", users).Find(&devices).Error
	return devices, err
}

func (s *Store) removeInvalid(ctx context.Context, device models.PushDevice) {
	// A slow FCM response must not erase a token that rotated in the meantime.
	_ = s.db.WithContext(ctx).Where("installation_id = ? AND user_id = ? AND token = ?",
		device.InstallationID, device.UserID, device.Token).Delete(&models.PushDevice{}).Error
}

type Handler struct{ store deviceStore }

func NewHandler(store *Store) *Handler { return &Handler{store: store} }

type deviceInput struct {
	InstallationID uuid.UUID `json:"installationId"`
	Token          string    `json:"token"`
	Platform       string    `json:"platform"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	input, ok := readDevice(w, r)
	if !ok {
		return
	}
	if input.Platform != "android" || !validToken(input.Token) {
		http.Error(w, "invalid Android push token", http.StatusBadRequest)
		return
	}
	device := models.PushDevice{
		InstallationID: input.InstallationID, UserID: userID, Token: input.Token, UpdatedAt: time.Now(),
	}
	if err := h.store.register(r.Context(), device); err != nil {
		http.Error(w, "could not register this device", http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Remove(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	input, ok := readDevice(w, r)
	if !ok {
		return
	}
	if err := h.store.remove(r.Context(), userID, input.InstallationID); err != nil {
		http.Error(w, "could not remove this device", http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func readDevice(w http.ResponseWriter, r *http.Request) (deviceInput, bool) {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
	decoder.DisallowUnknownFields()
	var input deviceInput
	var trailing json.RawMessage
	if decoder.Decode(&input) != nil || input.InstallationID == uuid.Nil || decoder.Decode(&trailing) != io.EOF {
		http.Error(w, "invalid push registration", http.StatusBadRequest)
		return deviceInput{}, false
	}
	return input, true
}

func validToken(token string) bool {
	return token != "" && len(token) <= 2048 && strings.IndexFunc(token, func(r rune) bool {
		return r < 0x21 || r > 0x7e
	}) < 0
}
