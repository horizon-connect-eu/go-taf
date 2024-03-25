package trustsource

import (
	"context"
	"gitlab-vs.informatik.uni-ulm.de/connect/taf-brussels-demo/pkg/trustassessment"
	"log"

	"gitlab-vs.informatik.uni-ulm.de/connect/taf-brussels-demo/pkg/message"
)

func Run(ctx context.Context,
	inputV2X chan message.InternalMessage,
	inputEvidenceCollection <-chan message.EvidenceCollectionMessage,
	output chan trustassessment.Command) {
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
		case received := <-inputV2X:
			if received.Rx == "TSM" {
				//log.Printf("I am TSM, received %+v\n", received)
				//output <- received
				cmd := trustassessment.CreateUpdateUpdateATOCommand("test", 4711)
				output <- cmd
			}

		case received := <-inputEvidenceCollection:
			log.Printf("[TSM], received %+v from evidence collection\n", received)
			//TODO: handle incoming evidence and generate update command
			cmd := trustassessment.CreateUpdateUpdateATOCommand("test", 4711)
			output <- cmd
		}

	}
}
