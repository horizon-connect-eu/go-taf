// # main package
//
// The main TAF application
package main

import (
	actualtlee "connect.informatik.uni-ulm.de/coordination/tlee-implementation/pkg/core"
	"context"
	"fmt"
	logging "github.com/vs-uulm/go-taf/internal/logger"
	"github.com/vs-uulm/go-taf/pkg/communication"
	"github.com/vs-uulm/go-taf/pkg/core"
	"github.com/vs-uulm/go-taf/pkg/crypto"
	"github.com/vs-uulm/go-taf/pkg/manager"
	internaltlee "github.com/vs-uulm/go-taf/pkg/tlee"
	"github.com/vs-uulm/go-taf/pkg/trustassessment"
	"github.com/vs-uulm/go-taf/pkg/trustmodel"
	"github.com/vs-uulm/taf-tlee-interface/pkg/tleeinterface"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vs-uulm/go-taf/pkg/config"
	"github.com/vs-uulm/go-taf/pkg/trustsource"
)

//go:generate go run ../plugins/plugins.go

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

	cryptoLib, err := crypto.NewCrypto(logging.CreateChildLogger(logger, "Crypto Library"), tafConfig.Crypto.KeyFolder, tafConfig.Crypto.Enabled)
	if err != nil {
		logger.Error("Error initializing crypto library")
	}

	tafContext := core.TafContext{
		Configuration: tafConfig,
		Logger:        logger,
		Context:       ctx,
		Identifier:    tafConfig.Identifier,
		Crypto:        cryptoLib,
	}

	//Channels
	tafChannels := core.TafChannels{
		TAMChannel:             make(chan core.Command, tafConfig.ChanBufSize),
		OutgoingMessageChannel: make(chan core.Message, tafConfig.ChanBufSize),
	}

	logger.Info("Starting TAF with ID '" + tafContext.Identifier + "'")

	communicationInterface, err := communication.NewInterface(tafContext, tafChannels)
	if err != nil {
		logger.Error("Error creating communication interface", err)
	}

	var tlee tleeinterface.TLEE
	if tafConfig.TLEE.UseInternalTLEE {
		tlee = &internaltlee.TLEE{}
	} else {
		tlee = &actualtlee.TLEE{}
	}

	trustAssessmentManager, err := trustassessment.NewManager(tafContext, tafChannels, tlee)
	if err != nil {
		logger.Error("Error creating TAM", err)
	}
	trustSourceManager, err := trustsource.NewManager(tafContext, tafChannels)
	if err != nil {
		logger.Error("Error creating TMM", err)
	}
	trustModelManager, err := trustmodel.NewManager(tafContext, tafChannels)
	if err != nil {
		logger.Error("Error creating TMM", err)
	}

	managers := manager.TafManagers{
		TSM: trustSourceManager,
		TAM: trustAssessmentManager,
		TMM: trustModelManager,
	}
	trustAssessmentManager.SetManagers(managers)
	trustModelManager.SetManagers(managers)
	trustSourceManager.SetManagers(managers)

	//Let's go
	go communicationInterface.Run()
	go trustAssessmentManager.Run()

	WaitForCtrlC()

}

// Blocks until the process receives SIGTERM (or equivalent).
func WaitForCtrlC() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}
