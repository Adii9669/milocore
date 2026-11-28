package services

import (
	"context"
	"testing"

	"chat-server/internals/db/models"

	"github.com/google/uuid"
)

type receiptRepo struct {
	MockMessageRepo
	updates int
	id      uuid.UUID
	reader  uuid.UUID
	read    bool
	changed bool
}

func (r *receiptRepo) UpdateDMReceipt(_ context.Context, id, reader uuid.UUID, read bool) (bool, error) {
	r.updates++
	r.id, r.reader, r.read = id, reader, read
	return r.changed, nil
}

func TestDMReceiptRequiresRecipientAndReturnsSenderOnChange(t *testing.T) {
	sender, recipient, id := uuid.New(), uuid.New(), uuid.New()
	message := &models.Message{ID: id, SenderID: sender, ReceiverID: &recipient}
	repo := &receiptRepo{changed: true}
	repo.FindMessageByIDFn = func(context.Context, uuid.UUID) (*models.Message, error) {
		return message, nil
	}
	service := NewMessageService(repo, nil, nil)

	for _, caller := range []uuid.UUID{sender, uuid.New()} {
		if _, err := service.RecordDMReceipt(context.Background(), id, caller, false); err == nil {
			t.Fatal("a user other than the recipient could mark a DM delivered")
		}
	}
	channel := uuid.New()
	message.ChannelID = &channel
	if _, err := service.RecordDMReceipt(context.Background(), id, recipient, true); err == nil {
		t.Fatal("a channel message accepted a DM receipt")
	}
	message.ChannelID = nil
	if repo.updates != 0 {
		t.Fatal("invalid receipt reached the repository")
	}

	got, err := service.RecordDMReceipt(context.Background(), id, recipient, true)
	if err != nil || got != sender || repo.id != id || repo.reader != recipient || !repo.read {
		t.Fatalf("valid receipt returned %v, %v with repository state %+v", got, err, repo)
	}
	repo.changed = false
	got, err = service.RecordDMReceipt(context.Background(), id, recipient, false)
	if err != nil || got != uuid.Nil {
		t.Fatalf("unchanged receipt should not notify sender: %v, %v", got, err)
	}
}
