package trustsource

import (
	"github.com/vs-uulm/go-taf/internal/flow/completionhandler"
	"github.com/vs-uulm/go-taf/pkg/command"
	"github.com/vs-uulm/go-taf/pkg/core"
	"github.com/vs-uulm/go-taf/pkg/trustmodel/session"
)

type Handler[CMD command.NotifyMessage] interface {
	AddSession(sess session.Session, handler *completionhandler.CompletionHandler)
	RemoveSession(sess session.Session, handler *completionhandler.CompletionHandler)
	TrustSourceType() core.TrustSource
	HandleNotify(cmd command.HandleNotify[CMD])
	RegisteredSessions() []string
}
