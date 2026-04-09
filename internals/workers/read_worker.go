package workers

import (
	"chat-server/internals/events"
	"chat-server/internals/services"
	"context"
	"log"
)

type ReadWorker struct {
	MessageService services.MessageService
	ReadChan       <-chan events.ReadEvent
}

func (w *ReadWorker) Read(ctx context.Context) {

	go func() {
		log.Println("📦 ReadWorker started")
		for {
			select {
			case event := <-w.ReadChan:
				err := w.MessageService.MarkRead(ctx, event.SenderID, event.ReaderID)
				if err != nil {
					log.Println("failed to mark Read:", err)
				}
			case <-ctx.Done():
				log.Println("🛑 ReadWorker stopped")
				return
			}
		}
	}()
}
