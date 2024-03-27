package evidencecollection

import (
	"context"

	"github.com/vs-uulm/go-taf/pkg/config"
	"github.com/vs-uulm/go-taf/pkg/message"
)

// Adapter is a function that writes Evidence to the supplied channel.
// It receives the following arguments: The context, an ID, the channel to write to, the configuration.
type Adapter func(context.Context, int, chan<- message.EvidenceCollectionMessage, config.Configuration)
