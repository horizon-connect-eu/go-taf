// # main package
//
// The main package is the entry point of the application. It is responsible for starting and stopping the application.
// Hello world
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gitlab-vs.informatik.uni-ulm.de/connect/taf-scalability-test/pkg/config"
	"gitlab-vs.informatik.uni-ulm.de/connect/taf-scalability-test/pkg/message"
	"gitlab-vs.informatik.uni-ulm.de/connect/taf-scalability-test/pkg/tam"
	"gitlab-vs.informatik.uni-ulm.de/connect/taf-scalability-test/pkg/tas"
	"gitlab-vs.informatik.uni-ulm.de/connect/taf-scalability-test/pkg/tmm"
	"gitlab-vs.informatik.uni-ulm.de/connect/taf-scalability-test/pkg/tmt"
	"gitlab-vs.informatik.uni-ulm.de/connect/taf-scalability-test/pkg/tsm"
	"gitlab-vs.informatik.uni-ulm.de/connect/taf-scalability-test/pkg/v2xlistener"
)

// Blocks until the process receives SIGTERM (or equivalent).
func waitForCtrlC() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}

// main is the entry point of the application.
// It starts all the components of the application and waits for a signal to stop the application.
func main() {
	tafConfig := config.DefaultConfig
	// First, see whether or not a config file path has been specified
	if filepath, ok := os.LookupEnv("TAF_CONFIG"); ok {
		var err error
		tafConfig, err = config.LoadJson(filepath)
		if err != nil {
			log.Fatalf("main: error reading config file %s: %s\n", filepath, err.Error())
		}
	}
	log.Printf("Running with configuration: %+v\n", tafConfig)

	c1 := make(chan message.Message, tafConfig.ChanBufSize)
	c2 := make(chan message.Message, tafConfig.ChanBufSize)

	c3 := make(chan message.Message, tafConfig.ChanBufSize)
	c4 := make(chan message.Message, tafConfig.ChanBufSize)

	c5 := make(chan message.TasResponse, tafConfig.ChanBufSize)
	c6 := make(chan message.TasQuery, tafConfig.ChanBufSize)

	tmts := map[string]int{}
	tmt.ParseXmlFiles("tmt/", tmts)

	ctx := context.Background()
	ctx, cancelFunc := context.WithCancel(ctx)
	defer time.Sleep(1 * time.Second) // TODO replace this cleanup interval with waitgroups
	defer cancelFunc()

	go v2xlistener.Run(ctx, tafConfig.V2XConfig, []chan message.Message{c1, c2})
	go tam.Run(ctx, tmts, c3, c4, c6, c5)

	go tmm.Run(ctx, c1, c3)
	go tsm.Run(ctx, c2, c4)

	go tas.Run(ctx, c5, c6)

	waitForCtrlC()

}
