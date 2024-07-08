// # main package
//
// The main TAF application
package main

import (
	"context"
	"crypto-library-interface/pkg/crypto"
	"fmt"
	logging "github.com/vs-uulm/go-taf/internal/logger"
	"github.com/vs-uulm/go-taf/pkg/communication"
	"github.com/vs-uulm/go-taf/pkg/core"
	"github.com/vs-uulm/go-taf/pkg/trustassessment"
	"log"
	"log/slog"
	"math/rand/v2"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vs-uulm/go-taf/pkg/config"
	"github.com/vs-uulm/go-taf/pkg/trustmodel"
	"github.com/vs-uulm/go-taf/pkg/trustsource"
)

//go:generate go run ../plugins/plugins.go

// Blocks until the process receives SIGTERM (or equivalent).
func WaitForCtrlC() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}

/*
The main TAF application that starts all the components of the application and waits for a signal to stop the application.
*/
func main() {
	tafConfig := config.DefaultConfig
	// First, see whether a config file path has been specified
	if filepath, ok := os.LookupEnv("TAF_CONFIG"); ok {
		var err error
		tafConfig, err = config.LoadJSON(filepath)
		if err != nil {
			log.Fatalf("main: error reading config file %s: %s\n", filepath, err.Error())
		}
	}

	logger := logging.CreateMainLogger(tafConfig.Logging)
	logger.Info("Configuration loaded")
	logger.Debug("Running with following configuration",
		slog.String("CONFIG", fmt.Sprintf("%+v", tafConfig)))

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer time.Sleep(1 * time.Second) // TODO: replace this cleanup interval with waitgroups
	defer cancelFunc()

	tafId := fmt.Sprintf("taf-%000000d", rand.IntN(999999))

	crypto.Init()

	tafContext := core.RuntimeContext{
		Configuration: tafConfig,
		Logger:        logger,
		Context:       ctx,
		Identifier:    tafId,
	}

	//Channels
	tafChannels := core.TafChannels{
		TAMChan:                make(chan core.Command, tafConfig.ChanBufSize),
		TSMChan:                make(chan core.Command, tafConfig.ChanBufSize),
		TMMChan:                make(chan core.Command, tafConfig.ChanBufSize),
		OutgoingMessageChannel: make(chan core.Message, tafConfig.ChanBufSize),
	}

	logger.Info("Starting TAF with ID " + tafId)

	communicationInterface, err := communication.NewInterface(tafContext, tafChannels)
	if err != nil {
		logger.Error("Error creating communication interface", err)
	}

	trustAssessmentManager, err := trustassessment.NewManager(tafContext, tafChannels)
	if err != nil {
		logger.Error("Error creating TAM", err)
	}

	trustModelManager, err := trustmodel.NewManager(tafContext, tafChannels)
	if err != nil {
		logger.Error("Error creating TMM", err)
	}

	trustSourceManager, err := trustsource.NewManager(tafContext, tafChannels)
	if err != nil {
		logger.Error("Error creating TMM", err)
	}

	//Let's go
	go communicationInterface.Run()
	go trustAssessmentManager.Run()
	go trustModelManager.Run()
	go trustSourceManager.Run()

	WaitForCtrlC()

}
