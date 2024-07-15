package trustsource

import (
	"context"
	logging "github.com/vs-uulm/go-taf/internal/logger"
	"github.com/vs-uulm/go-taf/pkg/command"
	"github.com/vs-uulm/go-taf/pkg/core"
	aivmsg "github.com/vs-uulm/go-taf/pkg/message/aiv"
	"log/slog"
)

type trustSourceManager struct {
	tafContext core.TafContext
	channels   core.TafChannels
	logger     *slog.Logger
}

func NewManager(tafContext core.TafContext, channels core.TafChannels) (trustSourceManager, error) {
	tsm := trustSourceManager{
		tafContext: tafContext,
		channels:   channels,
		logger:     logging.CreateChildLogger(tafContext.Logger, "TSM"),
	}
	return tsm, nil
}

func (tsm *trustSourceManager) Run() {
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
			if len(tsm.channels.TMMChan) != 0 || len(tsm.channels.TAMChan) != 0 || len(tsm.channels.TSMChan) != 0 {
				continue
			}
			return
		case incomingCmd := <-tsm.channels.TSMChan:

			switch cmd := incomingCmd.(type) {
			case command.HandleResponse[aivmsg.AivResponse]:
				tsm.handleAivResponse(cmd)
			default:
				tsm.logger.Warn("Command with no associated handling logic received by TSM", "Command Type", cmd.Type())
			}
		}
	}
}

func (t *trustSourceManager) handleAivResponse(cmd command.HandleResponse[aivmsg.AivResponse]) {
	t.logger.Info("TODO: handle AIV_RESPONSE: " + cmd.Response.AivEvidence.KeyRef)
}
