package listener

import (
	"github.com/horizon-connect-eu/go-taf/pkg/core"
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
	Version   int
}

func NewATLUpdatedEvent(fullTMI string, version int, oldATLs core.AtlResultSet, newATLs core.AtlResultSet) ATLUpdatedEvent {
	return ATLUpdatedEvent{
		Timestamp: time.Now(),
		EventType: ATL_UPDATED,
		Version:   version,
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
