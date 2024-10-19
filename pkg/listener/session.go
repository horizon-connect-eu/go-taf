package listener

import (
	"github.com/vs-uulm/go-taf/pkg/core"
	"time"
)

type SessionListener interface {
	OnSessionCreated(event SessionCreatedEvent)
	OnSessionTorndown(event SessionDeletedEvent)
}

type SessionCreatedEvent struct {
	Timestamp          time.Time
	SessionID          string
	TrustModelTemplate core.TrustModelTemplate
	ClientID           string
}

type SessionDeletedEvent struct {
	Timestamp          time.Time
	SessionID          string
	TrustModelTemplate core.TrustModelTemplate
	ClientID           string
}
