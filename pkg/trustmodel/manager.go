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

func (tmm trustModelManager) Run() {
	// Cleanup function:
	defer func() {
		tmm.logger.Info("Shutting down")
	}()

	/*
		// Create single TMI
		cmd := trustassessment.CreateInitTMICommand("demoModel", 1139)

		// Send initialization message to TAM
		output <- cmd
	*/

	// Do nothing until end
	for {
		// Each iteration, check whether we've been cancelled.
		if err := context.Cause(tmm.tafContext.Context); err != nil {
			return
		}
		select {
		case <-tmm.tafContext.Context.Done():
			return

		}
	}
}
