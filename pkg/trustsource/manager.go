package trustsource

import (
	"context"

	"gitlab-vs.informatik.uni-ulm.de/connect/taf-brussels-demo/pkg/message"
)

func Run(ctx context.Context, input chan message.InternalMessage, output chan message.InternalMessage) {
	// Cleanup function:
	defer func() {
		//log.Println("TSM: shutting down")
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
			if received.Rx == "TSM" {
				//log.Printf("I am TSM, received %+v\n", received)
				output <- received
			}
		}
	}
}
