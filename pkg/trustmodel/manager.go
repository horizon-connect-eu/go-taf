package trustmodel

import (
	"context"
	logging "github.com/vs-uulm/go-taf/internal/logger"
	"github.com/vs-uulm/go-taf/pkg/core"
	"log/slog"
)

type trustModelManager struct {
	tafContext core.RuntimeContext
	channels   core.TafChannels
	logger     *slog.Logger
}

func NewManager(tafContext core.RuntimeContext, channels core.TafChannels) (trustModelManager, error) {
	tmm := trustModelManager{
		tafContext: tafContext,
		channels:   channels,
		logger:     logging.CreateChildLogger(tafContext.Logger, "TMM")}
	return tmm, nil
}

func (tmm *trustModelManager) Run() {
	// Cleanup function:
	defer func() {
		tmm.logger.Info("Shutting down")
	}()

	for {
		// Each iteration, check whether we've been cancelled.
		if err := context.Cause(tmm.tafContext.Context); err != nil {
			return
		}
		select {
		case <-tmm.tafContext.Context.Done():
			if len(tmm.channels.TMMChan) != 0 || len(tmm.channels.TAMChan) != 0 || len(tmm.channels.TSMChan) != 0 {
				continue
			}
			return
		case incomingCmd := <-tmm.channels.TMMChan:

			switch cmd := incomingCmd.(type) {

			//case command.HandleResponse[aivmsg.AivResponse]:
			//	tmm.handleAivResponse(cmd)
			default:
				tmm.logger.Warn("Command with no associated handling logic received by TMM", "Command Type", cmd.Type())
			}
		}
	}
}
