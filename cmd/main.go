// # main package
//
// The main TAF application
package main

import (
	"context"
	"fmt"
	"github.com/vs-uulm/go-taf/cmd/flags"
	logging "github.com/vs-uulm/go-taf/internal/logger"
	"github.com/vs-uulm/go-taf/internal/version"
	"github.com/vs-uulm/go-taf/pkg/communication"
	"github.com/vs-uulm/go-taf/pkg/core"
	"github.com/vs-uulm/go-taf/pkg/crypto"
	"github.com/vs-uulm/go-taf/pkg/manager"
	"github.com/vs-uulm/go-taf/pkg/trustassessment"
	"github.com/vs-uulm/go-taf/pkg/trustmodel"
	"github.com/vs-uulm/go-taf/pkg/web"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vs-uulm/go-taf/pkg/config"
	"github.com/vs-uulm/go-taf/pkg/trustsource"
)

//go:generate go run ../plugins/trustmodels/updatetrustmodelhashes.go
//go:generate go run ../plugins/plugins.go

var (
	Version = version.Version
	Build   = version.Build
)

/*
The main TAF application that starts all the components of the application and waits for a signal to stop the application.
*/
func main() {
	tafConfig := config.DefaultConfig
	// check for the config file in environment and load, otherwise use default config
	if filepath, ok := os.LookupEnv("TAF_CONFIG"); ok {
		var err error
		tafConfig, err = config.LoadJSON(filepath)
		if err != nil {
			log.Fatalf("main: error reading config file %s: %s\n", filepath, err.Error())
		}
	}

	logger := logging.CreateMainLogger(tafConfig.Logging)
	logger.Info("Starting Standalone Trust Assessment Framework", "Version", Version, "Build", Build, "WEB-UI Flag", flags.WEB_UI)
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
		logger.Error("Error creating communication interface", "Error", err)
		os.Exit(-1)
		return
	}

	/*
		var tlee tleeinterface.TLEE
		if tafConfig.TLEE.UseInternalTLEE {
			tlee = internaltlee.SpawnNewTLEE(logging.CreateChildLogger(tafContext.Logger, "Internal TLEE"), tafConfig.TLEE.FilePath, tafConfig.TLEE.DebuggingMode)
		} else {
			tlee = actualtlee.SpawnNewTLEE(logging.CreateChildLogger(tafContext.Logger, "TLEE"), tafConfig.TLEE.FilePath, tafConfig.TLEE.DebuggingMode)
		}
	*/

	trustAssessmentManager, err := trustassessment.NewManager(tafContext, tafChannels)
	if err != nil {
		logger.Error("Error creating TAM", "Error", err)
		os.Exit(-1)
		return
	}
	trustSourceManager, err := trustsource.NewManager(tafContext, tafChannels)
	if err != nil {
		logger.Error("Error creating TMM", "Error", err)
		os.Exit(-1)
		return
	}
	trustModelManager, err := trustmodel.NewManager(tafContext, tafChannels)
	if err != nil {
		logger.Error("Error creating TMM", "Error", err)
		os.Exit(-1)
		return
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

	if flags.WEB_UI {
		server, err := web.New(tafContext)
		if err != nil {
			logger.Error("Error creating webserver", "Error", err)
			os.Exit(-1)
			return
		}
		server.SetManagers(managers)
		go server.Run()
		trustAssessmentManager.AddSessionListener(server)
		trustAssessmentManager.AddATLListener(server)
		trustAssessmentManager.AddTMIListener(server)
	}

	WaitForCtrlC()
}

/*
WaitForCtrlC blocks until the process receives SIGTERM (or equivalent).
*/
func WaitForCtrlC() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}
