package trustmodel

import (
	"context"
	"fmt"
	"gitlab-vs.informatik.uni-ulm.de/connect/taf-brussels-demo/pkg/trustassessment"
)

func Run(ctx context.Context, output chan trustassessment.Command) {
	// Cleanup function:
	defer func() {
		//log.Println("TMM: shutting down")
	}()

	//create single TMI
	cmd := trustassessment.CreateInitTMICommand("demoModel", 0x0fc9)
	fmt.Print(cmd.GetType())

	// Send initialization message to TAM
	output <- cmd

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
