package handlers

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

type TAMAccess interface {
	Sessions() map[string]session.Session
	DispatchToWorker(session session.Session, tmiID string, cmd core.Command)
}

type TSMAccess interface {
	SubscribeMBD(handler *completionhandler.CompletionHandler)
	UnsubscribeMBD(subID string, handler *completionhandler.CompletionHandler)

	SubscribeAIV(handler *completionhandler.CompletionHandler, sess session.Session)
	UnsubscribeAIV(subID string, handler *completionhandler.CompletionHandler)
}

type SubscriptionState uint16

const (
	NA SubscriptionState = iota
	SUBSCRIBING
	SUBSCRIBED
	UNSUBSCRIBING
)
