package v2xlistener

import (
	"context"
	"log"
	"time"

	"gitlab-vs.informatik.uni-ulm.de/connect/taf-scalability-test/pkg/message"
)

func Run(ctx context.Context, outputs []chan message.Message) {
	ticker := time.NewTicker(200 * time.Millisecond)

	defer func() {
		log.Println("V2XListener: shutting down")
		ticker.Stop()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			msg := message.Generate()
			// TODO: think about what should happen if a channel is full.
			// What if one is full and the other is not?
			// We should document this.
			for _, channel := range outputs {
				channel <- msg
			}
		}
	}
}
