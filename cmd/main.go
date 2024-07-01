// # main package
//
// The main TAF application
package main

import (
	"context"
	"fmt"
	logging "github.com/vs-uulm/go-taf/internal/logger"
	"github.com/vs-uulm/go-taf/pkg/communication"
	"github.com/vs-uulm/go-taf/pkg/core"
	"log"
	"log/slog"
	"math/rand/v2"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vs-uulm/go-taf/pkg/config"
	"github.com/vs-uulm/go-taf/pkg/message"
	"github.com/vs-uulm/go-taf/pkg/trustassessment"
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
The main TAF application thats tarts all the components of the application and waits for a signal to stop the application.
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
	tafContext := core.RuntimeContext{
		Configuration: tafConfig,
		Logger:        logger,
		Context:       ctx,
		Identifier:    tafId,
	}

	logger.Info("Starting TAF with ID " + tafId)

	incomingMessageChannel := make(chan communication.Message, tafConfig.ChanBufSize)
	outgoingMessageChannel := make(chan communication.Message, tafConfig.ChanBufSize)

	communicationInterface, err := communication.New(tafContext, incomingMessageChannel, outgoingMessageChannel)
	if err != nil {
		logger.Error("Error creating communication interface", err)
	}
	communicationInterface.Run(tafContext)

	/*
		time.Sleep(5 * time.Second)
		outgoingMessageChannel <- communication.NewMessage([]byte("{  \"sender\": \"a77b29bac8f1-taf\",  \"serviceType\": \"TAS\",  \"messageType\": \"TAS_INIT_REQUEST\",  \"responseId\": \"4c54a50f8e43\",  \"message\" : {  \"trustModelTemplate\":\"TRUSTMODEL@0.0.1\"}}"), "", "taf")
		outgoingMessageChannel <- communication.NewMessage([]byte("{  \"sender\": \"a77b29bac8f2-taf\",  \"serviceType\": \"TAS\",  \"messageType\": \"TAS_INIT_REQUEST\",  \"responseId\": \"4c54a50f8e43\",  \"message\" : {  \"trustModelTemplate\":\"TRUSTMODEL@0.0.1\"}}"), "", "taf")
	*/
	WaitForCtrlC()

	//	return

	//Create main channels
	//c1 := make(chan message.InternalMessage, tafConfig.ChanBufSize)
	c2 := make(chan message.InternalMessage, tafConfig.ChanBufSize)

	//c3 := make(chan message.InternalMessage, tafConfig.ChanBufSize)
	//c4 := make(chan message.InternalMessage, tafConfig.ChanBufSize)

	tmm2tamChannel := make(chan trustassessment.Command, tafConfig.ChanBufSize)
	tsm2tamChannel := make(chan trustassessment.Command, tafConfig.ChanBufSize)

	tmts := map[string]int{}

	//	go v2xlistener.Run(ctx, tafConfig.V2X, []chan message.InternalMessage{c1, c2})

	trustAssessmentManager, err := trustassessment.NewManager(tafContext, tmts)
	if err != nil {
		//LOG: log.Fatal(err)
	}
	go trustAssessmentManager.Run(tmm2tamChannel, tsm2tamChannel)

	go trustmodel.Run(ctx, tmm2tamChannel)
	go trustsource.Run(ctx, c2, tsm2tamChannel)

	WaitForCtrlC()

}
