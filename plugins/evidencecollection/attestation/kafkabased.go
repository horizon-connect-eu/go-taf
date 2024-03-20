package attestation

import (
	"gitlab-vs.informatik.uni-ulm.de/connect/taf-brussels-demo/pkg/config"
	"gitlab-vs.informatik.uni-ulm.de/connect/taf-brussels-demo/pkg/evidencecollection"
)

func init() {
	evidencecollection.RegisterEvidenceCollectionAdapter("kafkabased", NewKafkaBasedAttestation)
}

type KafkaBasedAttestation struct {
	Attestation
	config config.Configuration
}

func NewKafkaBasedAttestation(configuration config.Configuration) KafkaBasedAttestation {
	return KafkaBasedAttestation{}
}
