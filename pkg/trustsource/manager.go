package trustsource

import (
	"context"
	logging "github.com/vs-uulm/go-taf/internal/logger"
	"github.com/vs-uulm/go-taf/pkg/core"
	"log/slog"
)

type trustSourceManager struct {
	tafContext core.RuntimeContext
	channels   core.TafChannels
	logger     *slog.Logger
}

func NewManager(tafContext core.RuntimeContext, channels core.TafChannels) (trustSourceManager, error) {
	tsm := trustSourceManager{
		tafContext: tafContext,
		channels:   channels,
		logger:     logging.CreateChildLogger(tafContext.Logger, "TSM"),
	}
	return tsm, nil
}

func (tsm trustSourceManager) Run() {
	// Cleanup function:
	defer func() {
		tsm.logger.Info("Shutting down")
	}()

	for {
		// Each iteration, check whether we've been cancelled.
		if err := context.Cause(tsm.tafContext.Context); err != nil {
			return
		}
		select {
		case <-tsm.tafContext.Context.Done():
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
