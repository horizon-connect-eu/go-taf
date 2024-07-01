package main

import (
	"context"
	"encoding/csv"
	"fmt"
	logging "github.com/vs-uulm/go-taf/internal/logger"
	"github.com/vs-uulm/go-taf/internal/projectpath"
	"github.com/vs-uulm/go-taf/pkg/communication"
	"github.com/vs-uulm/go-taf/pkg/config"
	"github.com/vs-uulm/go-taf/pkg/core"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"time"
)

import _ "github.com/vs-uulm/go-taf/plugins/communication/filebased"
import _ "github.com/vs-uulm/go-taf/plugins/communication/kafkabased"
import _ "github.com/vs-uulm/go-taf/plugins/tam/add"
import _ "github.com/vs-uulm/go-taf/plugins/tam/mult"

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

	communicationInterface, err := communication.NewWithHandler(tafContext, nil, outgoingMessageChannel, "kafka-based")
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	communicationInterface.Run(tafContext)

	events, err := ReadFiles(filepath.FromSlash(testcase), logger)
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

func ReadFiles(pathDir string, logger *slog.Logger) ([]Event, error) {
	csvFile, err := os.Open(pathDir + "/script.csv")
	if err != nil {
		return nil, err
	}
	defer csvFile.Close()
	csvReader := csv.NewReader(csvFile)

	rawEvents, err := csvReader.ReadAll()
	events := make([]Event, 0)

	if err != nil {
		log.Fatal(err)
	}
	for lineNr, rawEvent := range rawEvents {
		timestamp, err := strconv.Atoi(rawEvent[0])
		if err != nil {
			logger.Error(fmt.Sprintf("error reading delay in line %d (%s): %+v", lineNr, rawEvent[0], err))
		}
		event := Event{}
		kafkaTopic := rawEvent[1]
		messagePath := rawEvent[2]

		message, err := os.ReadFile(pathDir + "/" + messagePath) // just pass the file name
		if err != nil {
			logger.Error(fmt.Sprintf("Error reading file '%s': %s", messagePath, err.Error()))
		}

		event.Timestamp = timestamp
		event.Topic = kafkaTopic
		event.Path = messagePath
		event.Message = message
		events = append(events, event)
	}

	// Sort messages by timestamp
	slices.SortFunc(events, func(a, b Event) int { return a.Timestamp - b.Timestamp })
	return events, nil
}

type Event struct {
	Timestamp int
	Topic     string
	Path      string
	Message   []byte
}
