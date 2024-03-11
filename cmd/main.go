package main

import (
	"bufio"
	"context"
	"os"
	"time"

	"gitlab-vs.informatik.uni-ulm.de/connect/taf-scalability-test/pkg/message"
	"gitlab-vs.informatik.uni-ulm.de/connect/taf-scalability-test/pkg/tam"
	"gitlab-vs.informatik.uni-ulm.de/connect/taf-scalability-test/pkg/tmm"
	"gitlab-vs.informatik.uni-ulm.de/connect/taf-scalability-test/pkg/tsm"
	"gitlab-vs.informatik.uni-ulm.de/connect/taf-scalability-test/pkg/v2xlistener"
)

func main() {
	c1 := make(chan message.Message, 1_000)
	c2 := make(chan message.Message, 1_000)

	c3 := make(chan message.Message, 1_000)
	c4 := make(chan message.Message, 1_000)

	ctx := context.Background()
	ctx, cancelFunc := context.WithCancel(ctx)
	defer time.Sleep(1 * time.Second)
	defer cancelFunc()

	go v2xlistener.Run(ctx, []chan message.Message{c1, c2})
	go tam.Run(c3, c4)

	go tmm.Run(c1, c3)
	go tsm.Run(c2, c4)

	bufio.NewReader(os.Stdin).ReadString('\n')

}
