package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	logging "github.com/vs-uulm/go-taf/internal/logger"
	"github.com/vs-uulm/go-taf/pkg/config"
	"github.com/vs-uulm/go-taf/pkg/message"
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

	readFiles("./res/workloads")

}

func readEvents(path string) ([]Event, error) {
	csvFile, err := os.Open(path)
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
		delay, err := strconv.Atoi(rawEvent[0])
		if err != nil {
			log.Printf("filebased evidence collector plugin: error reading delay in line %d (%s): %+v", lineNr, rawEvent[0], err)
		}
		event := Event{}
		err = json.Unmarshal([]byte(rawEvent[1]), &event)
		if err != nil {
			log.Printf("filebased evidence collector plugin: error reading event in line %d (%s): %+v", lineNr, rawEvent[1], err)
		}

		event.Timestamp = delay
		events = append(events, event)
	}

	// Sort messages by timestamp
	slices.SortFunc(events, func(a, b message.EvidenceCollectionMessage) int { return a.Timestamp - b.Timestamp })
	return events, nil
}

func readFiles(pathTestCases string) {
	testcases, ok := os.ReadDir(pathTestCases)

	if ok != nil {
		log.Fatal(ok)
	}

	//iterate over all directories in the provided path -> Each directory should represent here one test case
	for _, e := range testcases {
		pathScript := pathTestCases + e.Name() + "script.csv"
		_, ok := os.Stat(pathScript)

		if ok != nil {
			continue
		}

		events, err := readEvents(filepath.FromSlash(pathScript))

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

}
