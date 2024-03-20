package evidencecollection

import (
	"gitlab-vs.informatik.uni-ulm.de/connect/taf-brussels-demo/pkg/config"
	"gitlab-vs.informatik.uni-ulm.de/connect/taf-brussels-demo/pkg/message"
)

type Adapter func(chan message.EvidenceCollectionMessage, config.Configuration)
