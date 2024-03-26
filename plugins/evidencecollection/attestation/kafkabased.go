package attestation

import (
	"context"
	"fmt"

	"gitlab-vs.informatik.uni-ulm.de/connect/taf-brussels-demo/pkg/config"
	"gitlab-vs.informatik.uni-ulm.de/connect/taf-brussels-demo/pkg/evidencecollection"
	"gitlab-vs.informatik.uni-ulm.de/connect/taf-brussels-demo/pkg/message"
)

func init() {
	evidencecollection.RegisterEvidenceCollectionAdapter("kafkabased", NewKafkaBasedAttestation)
}

func NewKafkaBasedAttestation(ctx context.Context, id int, channel chan<- message.EvidenceCollectionMessage, config config.Configuration) {
	fmt.Println("Hello World from KafkaBasedAttestation!")
}
