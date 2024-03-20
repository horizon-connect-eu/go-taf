package attestation

import (
	"gitlab-vs.informatik.uni-ulm.de/connect/taf-brussels-demo/pkg/config"
	"gitlab-vs.informatik.uni-ulm.de/connect/taf-brussels-demo/pkg/evidencecollection"
)

func init() {
	evidencecollection.RegisterEvidenceCollectionAdapter("filebased", NewFileBasedAttestation)
}

type FileBasedAttestation struct {
	Attestation
	config config.Configuration
}

func NewFileBasedAttestation(configuration config.Configuration) FileBasedAttestation {
	return FileBasedAttestation{}
}
