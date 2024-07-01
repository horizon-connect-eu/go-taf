package trustsource

import (
	"context"
	"github.com/vs-uulm/go-taf/pkg/command"
	"github.com/vs-uulm/go-taf/pkg/message"
)

func Run(ctx context.Context,
	inputV2X chan message.InternalMessage,
	output chan command.Command) {
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

			/*
				case received := <-inputEvidenceCollection:
					//LOG: log.Printf("[TSM], received %+v from evidence collection\n", received)
					//TODO: handle incoming evidence and generate update command
					cmd := trustassessment.CreateUpdateTOCommand(uint64(received.TrustModelID), "TAF", received.Trustee, received.TS_ID, received.Evidence)
					output <- cmd
			*/
		}

	}
}
