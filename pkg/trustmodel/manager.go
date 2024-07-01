package trustmodel

import (
	"context"
	"github.com/vs-uulm/go-taf/pkg/command"
)

func Run(ctx context.Context, output chan command.Command) {
	// Cleanup function:
	defer func() {
		//log.Println("TMM: shutting down")
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
		if err := context.Cause(ctx); err != nil {
			return
		}
		select {
		case <-ctx.Done():
			return

		}
	}
}
