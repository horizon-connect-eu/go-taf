package tmm

import (
	"context"
	"log"

	"gitlab-vs.informatik.uni-ulm.de/connect/taf-scalability-test/pkg/message"
)

func Run(ctx context.Context, input chan message.Message, output chan message.Message) {
	// Cleanup function:
	defer func() {
		log.Println("TMM: shutting down")
	}()

	for {
		// Each iteration, check whether we've been cancelled.
		if err := context.Cause(ctx); err != nil {
			return
		}
		select {
		case <-ctx.Done():
			return
		case received := <-input:
			if received.Rx == "TMM" {
				log.Printf("I am TMM, received %+v\n", received)
				output <- received
			}
		}
	}
}
