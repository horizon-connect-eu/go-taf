package attestation

import (
	"fmt"

	"gitlab-vs.informatik.uni-ulm.de/connect/taf-brussels-demo/pkg/config"
	"gitlab-vs.informatik.uni-ulm.de/connect/taf-brussels-demo/pkg/evidencecollection"
	"gitlab-vs.informatik.uni-ulm.de/connect/taf-brussels-demo/pkg/message"
)

func init() {
	evidencecollection.RegisterEvidenceCollectionAdapter("filebased", NewFileBasedAttestation)
}

func NewFileBasedAttestation(channel chan message.EvidenceCollectionMessage, configuration config.Configuration) {
	fmt.Println("Hello World from FileBasedAttestation!")
}
