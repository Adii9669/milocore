package events

import "github.com/google/uuid"

type ReadEvent struct {
	MessageID string
	ReaderID  uuid.UUID
	SenderID  uuid.UUID
}
