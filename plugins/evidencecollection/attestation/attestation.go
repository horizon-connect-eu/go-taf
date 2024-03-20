package attestation

import (
	"gitlab-vs.informatik.uni-ulm.de/connect/taf-brussels-demo/pkg/evidencecollection"
	"gitlab-vs.informatik.uni-ulm.de/connect/taf-brussels-demo/pkg/message"
)

type Attestation struct {
	evidencecollection.Adapter
	outputChannel chan message.EvidenceCollectionMessage
}
