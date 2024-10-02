package trustsourcehandler

import (
	"github.com/vs-uulm/go-taf/internal/flow/completionhandler"
	"github.com/vs-uulm/go-taf/pkg/command"
	"github.com/vs-uulm/go-taf/pkg/core"
	"github.com/vs-uulm/go-taf/pkg/trustmodel/session"
)

/*
Handler is a entity to handle a specified type of trust source. Depending on the concrete type, this might include
no subscription, creating a single subscription for all sessions, or one subscription for each session.
*/
type Handler[CMD command.NotifyMessage] interface {
	/*
		AddSession adds the session to the trust source Handler. If an asynchronous operation is needed,
		the CompletionHandler will be used.
	*/
	AddSession(sess session.Session, handler *completionhandler.CompletionHandler)

	/*
		RemoveSession removes the session from the trust source Handler. If an asynchronous operation is needed,
		the CompletionHandler will be used.
	*/
	RemoveSession(sess session.Session, handler *completionhandler.CompletionHandler)

	/*
		TrustSourceType return the type of core.TrustSource.
	*/
	TrustSourceType() core.TrustSource

	/*
		HandleNotify will be called whenever there is a notification message coming from the trust source.
	*/
	HandleNotify(cmd command.HandleNotify[CMD])

	/*
		RegisteredSessions lists all sessions currently regsitered for this handler.
	*/
	RegisteredSessions() []string
}

/*
TAMAccess provides limited access for the handler to required functions from the TAM.
*/
type TAMAccess interface {
	Sessions() map[string]session.Session
	DispatchToWorker(session session.Session, tmiID string, cmd core.Command)
}

/*
TSMAccess provides limited access for the handler to required functions from the TSM.
*/
type TSMAccess interface {
	SubscribeMBD(handler *completionhandler.CompletionHandler)
	UnsubscribeMBD(subID string, handler *completionhandler.CompletionHandler)

	SubscribeAIV(handler *completionhandler.CompletionHandler, sess session.Session)
	UnsubscribeAIV(subID string, handler *completionhandler.CompletionHandler)
}

/*
SubscriptionState specifies the state a handler is in.
*/
type SubscriptionState uint16

const (
	NA SubscriptionState = iota
	SUBSCRIBING
	SUBSCRIBED
	UNSUBSCRIBING
)
