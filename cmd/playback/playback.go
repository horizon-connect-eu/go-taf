package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	logging "github.com/vs-uulm/go-taf/internal/logger"
	"github.com/vs-uulm/go-taf/internal/util"
	"github.com/vs-uulm/go-taf/pkg/config"
	"github.com/vs-uulm/go-taf/pkg/core"
	"github.com/vs-uulm/go-taf/plugins/communication/kafkabased"
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
	//specification of userstory
	testcase := flag.String("story", "", "path to the directory with the storyline specification - should include a script.csv file and the single json messages")
	//specification of config path
	configPath := flag.String("config", "", "path to the file with the configuration specification")
	//specification of targets
	target := flag.Bool("target", false, "list of targets can be specified - if targets are provided, messages are send from the playback tool only to these targets (all other messages are filtered out)")

	flag.Parse()

	if *testcase == "" {
		fmt.Fprintln(os.Stderr, "Story parameter is missing - please use the story parameter to specify the directory of the story line")
		fmt.Fprintln(os.Stderr, "usage:   ./playback -story=path [-config=path] [-target target list]")
		fmt.Fprintln(os.Stderr, "example: ./playback -story=storydirectory/storyline1 -config=configdirectory/config1.json -target taf aiv mbd")
		os.Exit(1)
	}

	absPathTestCases := *testcase

	//crypto.Init()
	tafConfig := config.DefaultConfig

	if *configPath != "" {
		var err error
		tafConfig, err = config.LoadJSON(*configPath)

		if err != nil {
			//log.Fatalf("main: error reading config file %s: %s\n", configFile, err.Error())
			fmt.Fprintln(os.Stderr, "Config parameter is incorrect - specified file "+*configPath+" not found")
			fmt.Fprintln(os.Stderr, "usage:   ./playback -story=path [-config=path] [-target target list]")
			fmt.Fprintln(os.Stderr, "example: ./playback -story=storydirectory/storyline1 -config=configdirectory/config1.json -target taf aiv mbd")
			os.Exit(1)
		}
	} else if filepath, ok := os.LookupEnv("TAF_CONFIG"); ok {
		var err error
		tafConfig, err = config.LoadJSON(filepath)
		if err != nil {
			//log.Fatalf("main: error reading config file %s: %s\n", filepath, err.Error())
			fmt.Fprintln(os.Stderr, "Environment variable is incorrect - specified file "+filepath+" not found")
		}
	}

	var targetEntities []string
	if *target == true {
		targetEntities = flag.Args()
	}

	logger := logging.CreateMainLogger(tafConfig.Logging)
	logger.Info("Configuration loaded")
	logger.Debug("Running with following configuration",
		slog.String("CONFIG", fmt.Sprintf("%+v", tafConfig)))

	ctx, cancelFunc := context.WithCancel(context.Background())

	outgoingMessageChannel := make(chan core.Message, tafConfig.ChanBufSize)

	tafContext := core.RuntimeContext{
		Configuration: tafConfig,
		Logger:        logger,
		Context:       ctx,
		Identifier:    "playback",
	}

	//Dummy channel for received messages from the communication interface.
	//As we will receive (at least some of) the messages sent by ourselves, we consume and ignore them in a separate go-routine.
	incomingMessageChannel := make(chan core.Message, tafContext.Configuration.ChanBufSize)
	go func() {
		for {
			select {
			case msg := <-incomingMessageChannel:
				util.UNUSED(msg)
			}
		}
	}()
	//Directly create Kafka-based Communication Interface Handler.
	go kafkabased.NewKafkaBasedHandler(tafContext, incomingMessageChannel, outgoingMessageChannel)

	defer time.Sleep(1 * time.Second) // TODO: replace this cleanup interval with waitgroups
	defer cancelFunc()

	/*
		communicationInterface, err := communication.NewWithHandler(tafContext, nil, outgoingMessageChannel, "kafka-based")
		if err != nil {
			logger.Error(err.Error())
			os.Exit(1)
		}
		communicationInterface.Run(tafContext)
	*/

	events, err := ReadFiles(filepath.FromSlash(absPathTestCases), targetEntities, *target, logger)
	if err != nil {
		logger.Error(err.Error())
		fmt.Fprintln(os.Stderr, "Invalid input for the story parameter - Please make sure you enter a correct path and the directory contains a script.csv file")
		fmt.Fprintln(os.Stderr, "usage:   ./playback -story=path [-config=path] [-target targetlist]")
		fmt.Fprintln(os.Stderr, "example: ./playback -story=storydirectory/storyline1 -config=configdirectory/config1.json -target taf aiv mbd")
		os.Exit(1)
	}

	// send all messages at the appropriate time
	internalTime := 0
	for _, event := range events {
		var jsonMap map[string]interface{}
		json.Unmarshal(event.Message, &jsonMap)
		// Sleep until the next event is due

		/*
			TODO: fix undeliberate usage of evidence
			evidence, _ := crypto.GenerateEvidence()
			jsonMap["message"].(map[string]interface{})["evidence"] = evidence
		*/

		sleepFor := event.Timestamp - internalTime
		time.Sleep(time.Duration(sleepFor) * time.Millisecond)
		internalTime = event.Timestamp

		logger.Info(fmt.Sprintf("Sending message at timestamp %d ms to topic '%s'", event.Timestamp, event.Topic))
		event.Message, _ = json.Marshal(jsonMap)
		outgoingMessageChannel <- core.NewMessage(event.Message, event.Sender, event.Topic)
	}

	time.Sleep(time.Duration(5000) * time.Millisecond)

}

func ReadFiles(pathDir string, targetEntities []string, target bool, logger *slog.Logger) ([]Event, error) {
	csvFile, err := os.Open(pathDir + "/script.csv")
	if err != nil {
		return nil, err
	}
	defer csvFile.Close()
	csvReader := csv.NewReader(csvFile)

	rawEvents, err := csvReader.ReadAll()
	events := make([]Event, 0)

	if err != nil {
		//log.Fatal(err)
		return nil, err
	}
	for lineNr, rawEvent := range rawEvents {
		timestamp, err := strconv.Atoi(rawEvent[0])
		if err != nil {
			logger.Error(fmt.Sprintf("error reading delay in line %d (%s): %+v", lineNr, rawEvent[0], err))
		}
		event := Event{
			Timestamp: timestamp,
			Sender:    rawEvent[1],
			Topic:     rawEvent[2],
			Path:      rawEvent[3],
		}

		if target == true {
			if !checkStringInArray(event.Topic, targetEntities) {
				continue
			} else {
				sourceEntity := rawEvent[1]
				if checkStringInArray(sourceEntity, targetEntities) { // If source entity is also target entity, this entity is under test and will produce the messages on its own, therefore this message does not have to be replayed
					continue
				}
			}
		}

		message, err := os.ReadFile(pathDir + "/" + event.Path) // just pass the file name
		// str_message := string(message) // just pass the file name

		if err != nil {
			logger.Error(fmt.Sprintf("Error reading file '%s': %s", event.Path, err.Error()))
			return nil, err
		}

		event.Message = message
		events = append(events, event)
	}

	// Sort messages by timestamp
	slices.SortFunc(events, func(a, b Event) int { return a.Timestamp - b.Timestamp })
	return events, nil
}

func checkStringInArray(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

type Event struct {
	Timestamp int
	Sender    string
	Topic     string
	Path      string
	Message   []byte
}
