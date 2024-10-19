package listener

import (
	"github.com/vs-uulm/go-taf/pkg/core"
	"time"
)

type ActualTrustLevelListener interface {
	OnATLUpdated(event ATLUpdatedEvent)
	OnATLRemoved(event ATLRemovedEvent)
}

type ATLUpdatedEvent struct {
	Timestamp time.Time
	FullTMI   string
	OldATLs   core.AtlResultSet
	NewATLs   core.AtlResultSet
}

type ATLRemovedEvent struct {
	Timestamp time.Time
	FullTMI   string
}
