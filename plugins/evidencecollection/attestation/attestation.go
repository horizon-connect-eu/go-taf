package attestation

import (
	"github.com/vs-uulm/go-taf/pkg/evidencecollection"
	"github.com/vs-uulm/go-taf/pkg/message"
)

type Attestation struct {
	evidencecollection.Adapter
	outputChannel chan<- message.EvidenceCollectionMessage
}
