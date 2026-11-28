package chathistory

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"chat-server/internals/repository"
	"chat-server/internals/transport/dto"

	"github.com/google/uuid"
)

func historyQuery(r *http.Request) (int, *repository.HistoryCursor, error) {
	params := r.URL.Query()
	limit, err := strconv.Atoi(params.Get("limit"))
	if err != nil || limit < 1 || limit > 100 {
		limit = 50
	}
	raw := params.Get("cursor")
	if raw == "" {
		return limit, nil, nil
	}
	date, id, composite := strings.Cut(raw, "|")
	at, err := time.Parse(time.RFC3339Nano, date)
	if err != nil {
		return limit, nil, err
	}
	cursor := &repository.HistoryCursor{At: at}
	if composite {
		cursor.ID, err = uuid.Parse(id)
		if err != nil || cursor.ID == uuid.Nil {
			return limit, nil, fmt.Errorf("invalid cursor message ID")
		}
	}
	return limit, cursor, nil
}

func pageCursor(messages []dto.MessageResponse, hasMore, stable bool) *string {
	if !hasMore || len(messages) == 0 {
		return nil
	}
	oldest := messages[0]
	value := oldest.CreatedAt.Format(time.RFC3339Nano)
	if stable {
		value += "|" + oldest.ID
	}
	return &value
}
