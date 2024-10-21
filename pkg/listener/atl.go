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
	EventType EventType
	FullTMI   string
	OldATLs   core.AtlResultSet
	NewATLs   core.AtlResultSet
}

func NewATLUpdatedEvent(fullTMI string, oldATLs core.AtlResultSet, newATLs core.AtlResultSet) ATLUpdatedEvent {
	return ATLUpdatedEvent{
		Timestamp: time.Now(),
		EventType: ATL_UPDATED,
		FullTMI:   fullTMI,
		OldATLs:   oldATLs,
		NewATLs:   newATLs,
	}
}

func (e ATLUpdatedEvent) Event() EventType {
	return e.EventType
}

type ATLRemovedEvent struct {
	Timestamp time.Time
	EventType EventType
	FullTMI   string
}

func NewATLRemovedEvent(fullTMI string) ATLRemovedEvent {
	return ATLRemovedEvent{
		Timestamp: time.Now(),
		EventType: ATL_REMOVED,
		FullTMI:   fullTMI}
}

func (e ATLRemovedEvent) Event() EventType {
	return e.EventType
}
