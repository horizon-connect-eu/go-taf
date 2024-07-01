package main

import (
	"context"
	"encoding/csv"
	"fmt"
	logging "github.com/vs-uulm/go-taf/internal/logger"
	"github.com/vs-uulm/go-taf/pkg/config"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strconv"
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

	_, cancelFunc := context.WithCancel(context.Background())
	defer time.Sleep(1 * time.Second) // TODO: replace this cleanup interval with waitgroups
	defer cancelFunc()

	testcase := "example" //specification of testcase -> directory name in workloads folder

	sendMessages("./res/workloads" + "/" + testcase)

}

func readFiles(pathDir string) ([]Event, error) {
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
			log.Printf("filebased evidence collector plugin: error reading delay in line %d (%s): %+v", lineNr, rawEvent[0], err)
		}
		event := Event{}
		kafkaTopic := rawEvent[1]
		messagePath := rawEvent[2]

		message, err := os.ReadFile(pathDir + "/" + messagePath) // just pass the file name
		if err != nil {
			continue
		}

		event.Timestamp = timestamp
		event.Topic = kafkaTopic
		event.Path = messagePath
		event.Message = string(message)
		events = append(events, event)
	}

	// Sort messages by timestamp
	slices.SortFunc(events, func(a, b Event) int { return a.Timestamp - b.Timestamp })
	return events, nil
}

func sendMessages(pathScript string) {
	events, err := readFiles(filepath.FromSlash(pathScript))

	if err != nil {
		log.Fatal(err)
	}
	// send all messages at the appropriate time
	internalTime := 0
	for _, event := range events {
		// Sleep until the next event is due
		sleepFor := event.Timestamp - internalTime
		time.Sleep(time.Duration(sleepFor) * time.Millisecond)
		internalTime = event.Timestamp

		// Send the next event
		//LOG: log.Printf("filebased evidence collector plugin: sending %+v\n", event)
		//LOG: log.Printf("filebased evidence collector plugin: sent %+v\n", event)
	}
}
