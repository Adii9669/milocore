package repository

import (
	"strings"
	"testing"
	"time"

	"chat-server/internals/db/models"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestHistoryCursorBuildsTupleBoundaryForEqualTimestamps(t *testing.T) {
	db, err := gorm.Open(postgres.Open("host=localhost user=unused dbname=unused"), &gorm.Config{
		DisableAutomaticPing: true,
		DryRun:               true,
	})
	if err != nil {
		t.Fatal(err)
	}
	cursor := &HistoryCursor{At: time.Now(), ID: uuid.New()}
	query := beforeHistoryCursor(db.Model(&models.Message{}), cursor).Find(&[]models.Message{})
	if query.Error != nil {
		t.Fatal(query.Error)
	}
	args := query.Statement.Vars
	if !strings.Contains(query.Statement.SQL.String(), "(messages.created_at, messages.id) < ($1, $2)") ||
		len(args) != 2 || args[0] != cursor.At || args[1] != cursor.ID {
		t.Fatalf("unexpected stable query: %s, %v", query.Statement.SQL.String(), args)
	}
}
