package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gitlab-vs.informatik.uni-ulm.de/connect/taf-scalability-test/pkg/message"
	"gitlab-vs.informatik.uni-ulm.de/connect/taf-scalability-test/pkg/tam"
	"gitlab-vs.informatik.uni-ulm.de/connect/taf-scalability-test/pkg/tas"
	"gitlab-vs.informatik.uni-ulm.de/connect/taf-scalability-test/pkg/tmm"
	"gitlab-vs.informatik.uni-ulm.de/connect/taf-scalability-test/pkg/tsm"
	"gitlab-vs.informatik.uni-ulm.de/connect/taf-scalability-test/pkg/v2xlistener"
)

// Blocks until the process receives SIGTERM (or equivalent).
func waitForCtrlC() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}

func main() {

	c1 := make(chan message.Message, 1_000)
	c2 := make(chan message.Message, 1_000)

	c3 := make(chan message.Message, 1_000)
	c4 := make(chan message.Message, 1_000)

	c5 := make(chan message.TasResponse, 1_000)
	c6 := make(chan message.TasQuery, 1_000)

	ctx := context.Background()
	ctx, cancelFunc := context.WithCancel(ctx)
	defer time.Sleep(1 * time.Second) // TODO replace this cleanup interval with waitgroups
	defer cancelFunc()

	go v2xlistener.Run(ctx, []chan message.Message{c1, c2})
	go tam.Run(ctx, c3, c4, c6, c5)

	go tmm.Run(ctx, c1, c3)
	go tsm.Run(ctx, c2, c4)

	go tas.Run(ctx, c5, c6)

	waitForCtrlC()

}
