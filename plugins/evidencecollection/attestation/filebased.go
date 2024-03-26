package attestation

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"time"

	"gitlab-vs.informatik.uni-ulm.de/connect/taf-brussels-demo/pkg/config"
	"gitlab-vs.informatik.uni-ulm.de/connect/taf-brussels-demo/pkg/evidencecollection"
	"gitlab-vs.informatik.uni-ulm.de/connect/taf-brussels-demo/pkg/message"
)

func init() {
	evidencecollection.RegisterEvidenceCollectionAdapter("filebased", NewFileBasedAttestation)
}

func readEvents(path string) ([]message.EvidenceCollectionMessage, error) {
	csvFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer csvFile.Close()
	csvReader := csv.NewReader(csvFile)

	rawEvents, err := csvReader.ReadAll()
	events := make([]message.EvidenceCollectionMessage, 0)

	if err != nil {
		log.Fatal(err)
	}
	for lineNr, rawEvent := range rawEvents {
		delay, err := strconv.Atoi(rawEvent[0])
		if err != nil {
			log.Printf("filebased evidence collector plugin: error reading delay in line %d (%s): %+v", lineNr, rawEvent[0], err)
		}
		event := message.EvidenceCollectionMessage{}
		err = json.Unmarshal([]byte(rawEvent[1]), &event)
		if err != nil {
			log.Printf("filebased evidence collector plugin: error reading event in line %d (%s): %+v", lineNr, rawEvent[1], err)
		}

		event.Timestamp = delay // TODO why not have this as a json field in the first place?
		events = append(events, event)
	}

	// Sort messages by timestamp
	slices.SortFunc(events, func(a, b message.EvidenceCollectionMessage) int { return a.Timestamp - b.Timestamp })
	return events, nil
}

func NewFileBasedAttestation(ctx context.Context, id int, channel chan<- message.EvidenceCollectionMessage, configuration config.Configuration) {
	path, ok := configuration.EvidenceCollection.Adapters[id].Params["path"]
	if !ok {
		log.Fatal("filebased evidence collector plugin: no path specified to read from")
	}
	events, err := readEvents(filepath.FromSlash(path))
	log.Printf("filebased evidence collector plugin: read %d messages", len(events))

	if err != nil {
		log.Fatal(err)
	}
	// send all messages at the appropriate time
	internalTime := 0
	for _, event := range events {
		// Sleep until the next event is due
		sleepFor := event.Timestamp - internalTime
		time.Sleep(time.Duration(sleepFor) * time.Second)
		internalTime = event.Timestamp

		// Send the next event
		log.Printf("filebased evidence collector plugin: sending %+v\n", event)
		channel <- event
		log.Printf("filebased evidence collector plugin: sent %+v\n", event)
	}
}
