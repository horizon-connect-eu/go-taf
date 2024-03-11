package v2xlistener

import (
	"context"
	"fmt"
	"time"

	"gitlab-vs.informatik.uni-ulm.de/connect/taf-scalability-test/pkg/message"
)

func Run(ctx context.Context, outputs []chan message.Message) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("V2XListener: shutting down")
			return
		default:
			msg := message.Generate()
			for _, channel := range outputs {
				channel <- msg
			}
			time.Sleep(200 * time.Millisecond)
		}
	}
}
