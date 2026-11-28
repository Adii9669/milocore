package chathistory

import (
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"chat-server/internals/transport/dto"

	"github.com/google/uuid"
)

func TestHistoryCursorPreservesMessageIDAndLegacyTime(t *testing.T) {
	at := time.Date(2026, 11, 28, 8, 0, 0, 123456000, time.UTC)
	oldest := uuid.New()
	messages := []dto.MessageResponse{
		{ID: oldest.String(), CreatedAt: at},
		{ID: uuid.NewString(), CreatedAt: at},
	}
	for _, stable := range []bool{false, true} {
		value := pageCursor(messages, true, stable)
		request := httptest.NewRequest("GET", "/api/chats/dm/example?limit=0&cursor="+url.QueryEscape(*value), nil)
		limit, cursor, err := historyQuery(request)
		if err != nil || limit != 50 || cursor == nil || !cursor.At.Equal(at) {
			t.Fatalf("invalid round trip: %d, %+v, %v", limit, cursor, err)
		}
		if stable && cursor.ID != oldest || !stable && cursor.ID != uuid.Nil {
			t.Fatalf("unexpected cursor ID %s (stable %v)", cursor.ID, stable)
		}
	}
	request := httptest.NewRequest("GET", "/api/chats/dm/example?cursor=broken", nil)
	if _, _, err := historyQuery(request); err == nil {
		t.Fatal("malformed cursor was accepted")
	}
}
