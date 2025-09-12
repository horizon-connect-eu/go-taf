package listener

import (
	"github.com/horizon-connect-eu/go-taf/pkg/core"
	"time"
)

type SessionListener interface {
	OnSessionCreated(event SessionCreatedEvent)
	OnSessionTorndown(event SessionTorndownEvent)
}

type SessionCreatedEvent struct {
	Timestamp          time.Time
	EventType          EventType
	SessionID          string
	TrustModelTemplate string
	ClientID           string
}

func NewSessionCreatedEvent(sessionID string, trustModelTemplate core.TrustModelTemplate, clientID string) SessionCreatedEvent {
	return SessionCreatedEvent{
		Timestamp: time.Now(),
		EventType: SESSION_CREATED,
		SessionID: sessionID, TrustModelTemplate: trustModelTemplate.Identifier(), ClientID: clientID}
}
func (e SessionCreatedEvent) Event() EventType {
	return e.EventType
}

type SessionTorndownEvent struct {
	Timestamp          time.Time
	EventType          EventType
	SessionID          string
	TrustModelTemplate string
	ClientID           string
}

func NewSessionTorndownEvent(sessionID string, trustModelTemplate core.TrustModelTemplate, clientID string) SessionTorndownEvent {
	return SessionTorndownEvent{
		Timestamp: time.Now(),
		EventType: SESSION_TORNDOWN,
		SessionID: sessionID, TrustModelTemplate: trustModelTemplate.Identifier(), ClientID: clientID}
}
func (e SessionTorndownEvent) Event() EventType {
	return e.EventType
}
