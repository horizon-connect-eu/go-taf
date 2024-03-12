package v2xlistener

import (
	"context"
	"log"
	"time"

	"gitlab-vs.informatik.uni-ulm.de/connect/taf-scalability-test/pkg/config"
	"gitlab-vs.informatik.uni-ulm.de/connect/taf-scalability-test/pkg/message"
)

func Run(ctx context.Context, v2xconfig config.V2XConfiguration, outputs []chan message.Message) {
	ticker := time.NewTicker(time.Duration(v2xconfig.SendIntervalMs) * time.Millisecond)

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
