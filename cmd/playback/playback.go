package main

import (
	"context"
	"fmt"
	logging "github.com/vs-uulm/go-taf/internal/logger"
	"github.com/vs-uulm/go-taf/internal/projectpath"
	"github.com/vs-uulm/go-taf/pkg/communication"
	"github.com/vs-uulm/go-taf/pkg/config"
	"github.com/vs-uulm/go-taf/pkg/core"
	"github.com/vs-uulm/go-taf/plugins/communication/filebased"
	"github.com/vs-uulm/go-taf/plugins/communication/kafkabased"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

/*
A helper command to play back test workloads via Kafka.
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

	//specification of testcase -> directory name in workloads folder
	testcase := projectpath.Root + "/res/workloads/example" //TODO: make CLI flag: https://gobyexample.com/command-line-flags

	outgoingMessageChannel := make(chan communication.Message, tafConfig.ChanBufSize)

	tafContext := core.RuntimeContext{
		Configuration: tafConfig,
		Logger:        logger,
		Context:       ctx,
		Identifier:    "playback",
	}

	go kafkabased.NewKafkaBasedHandler(tafContext, nil, outgoingMessageChannel)

	events, err := filebased.ReadFiles(filepath.FromSlash(testcase), logger)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	// send all messages at the appropriate time
	internalTime := 0
	for _, event := range events {
		// Sleep until the next event is due
		sleepFor := event.Timestamp - internalTime
		time.Sleep(time.Duration(sleepFor) * time.Millisecond)
		internalTime = event.Timestamp

		logger.Info(fmt.Sprintf("Sending message at timestamp %d ms to topic '%s'", event.Timestamp, event.Topic))

		outgoingMessageChannel <- communication.NewMessage(event.Message, "", event.Topic)
	}

}
