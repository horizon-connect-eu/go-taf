package filebased

import (
	"encoding/csv"
	"fmt"
	logging "github.com/vs-uulm/go-taf/internal/logger"
	"github.com/vs-uulm/go-taf/internal/projectpath"
	"github.com/vs-uulm/go-taf/pkg/communication"
	"github.com/vs-uulm/go-taf/pkg/core"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"time"
)

func init() {
	communication.RegisterCommunicationHandler("file-based", NewFileBasedHandler)
}

func NewFileBasedHandler(tafContext core.RuntimeContext, inboxChannel chan<- core.Message, outboxChannel <-chan core.Message) {
	logger := logging.CreateChildLogger(tafContext.Logger, "File Communication Handler")
	logger.Info("Starting file-based communication handler.")

	go handleOutgoingMessages(tafContext, logger, outboxChannel)
	go handleIncomingMessages(tafContext, logger, inboxChannel)
}

/*
Print message content to console.
*/
func handleOutgoingMessages(tafContext core.RuntimeContext, logger *slog.Logger, outboxChannel <-chan core.Message) {
	for {
		select {
		case msg := <-outboxChannel:
			logger.Info(fmt.Sprintf("Outgoing message from %s to %s:", msg.Source(), msg.Destination()))
		}
	}
}

func handleIncomingMessages(tafContext core.RuntimeContext, logger *slog.Logger, inboxChannel chan<- core.Message) {

	testcase := projectpath.Root + "/res/workloads/example" //TODO: make CLI flag: https://gobyexample.com/command-line-flags

	events, err := ReadFiles(filepath.FromSlash(testcase), logger)
	if err != nil {
		logger.Error(err.Error())
	}

	// send all messages at the appropriate time
	internalTime := 0
	for _, event := range events {
		// Sleep until the next event is due
		sleepFor := event.Timestamp - internalTime
		time.Sleep(time.Duration(sleepFor) * time.Millisecond)
		internalTime = event.Timestamp

		logger.Info(fmt.Sprintf("Sending message at timestamp %d ms to topic '%s'", event.Timestamp, event.Topic))

		inboxChannel <- core.NewMessage(event.Message, "", event.Topic)
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
