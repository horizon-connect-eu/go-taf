package attestation

import (
	"context"
	"fmt"

	"github.com/vs-uulm/go-taf/pkg/config"
	"github.com/vs-uulm/go-taf/pkg/evidencecollection"
	"github.com/vs-uulm/go-taf/pkg/message"
)

func init() {
	evidencecollection.RegisterEvidenceCollectionAdapter("kafkabased", NewKafkaBasedAttestation)
}

func NewKafkaBasedAttestation(ctx context.Context, id int, channel chan<- message.EvidenceCollectionMessage, config config.Configuration) {
	fmt.Println("Hello World from KafkaBasedAttestation!")
}
