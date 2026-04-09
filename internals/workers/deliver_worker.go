package workers

import (
	"chat-server/internals/services"
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
)

type DeliveryWorker struct {
	MessageService services.MessageService
	DeliveredChan  <-chan string
}

func (w *DeliveryWorker) Start(ctx context.Context) {

	go func() {
		log.Println("📦 DeliveryWorker started")
		for {
			select {
			case msgID := <-w.DeliveredChan:
				msg, err := uuid.Parse(msgID)
				if err != nil {
					fmt.Printf("parsing failed")
					return
				}
				err = w.MessageService.MarkDelivered(ctx, msg)
				if err != nil {
					log.Println("failed to mark delivered:", err)
				}
			case <-ctx.Done():
				log.Println("🛑 DeliveryWorker stopped")
				return
			}
		}
	}()
}
