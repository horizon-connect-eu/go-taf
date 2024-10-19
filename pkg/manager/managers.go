package manager

import (
	"github.com/vs-uulm/go-taf/internal/flow/completionhandler"
	"github.com/vs-uulm/go-taf/pkg/command"
	"github.com/vs-uulm/go-taf/pkg/core"
	"github.com/vs-uulm/go-taf/pkg/listener"
	messages "github.com/vs-uulm/go-taf/pkg/message"
	aivmsg "github.com/vs-uulm/go-taf/pkg/message/aiv"
	mbdmsg "github.com/vs-uulm/go-taf/pkg/message/mbd"
	tasmsg "github.com/vs-uulm/go-taf/pkg/message/tas"
	tchmsg "github.com/vs-uulm/go-taf/pkg/message/tch"
	v2xmsg "github.com/vs-uulm/go-taf/pkg/message/v2x"
	"github.com/vs-uulm/go-taf/pkg/trustmodel/session"
)

type TafManagers struct {
	TSM TrustSourceManager
	TAM TrustAssessmentManager
	TMM TrustModelManager
}

/*
The CooperativeManager is a manager type that knows about other managers (via SetManagers) and can thus call their
functions .
*/
type CooperativeManager interface {
	SetManagers(managers TafManagers)
}

/*
The RunnableManager is a manager with a Run method that must be called after manager initialization via a go-routine.
*/
type RunnableManager interface {
	Run()
}

/*
The TrustAssessmentManager is an internal component responsible for handling communication with client applications and
dispatching operations to the TSM and TMM. It is running in a dedicated go-routine with an exclusive channel that
contains incoming messages and updates operations either to be handled by the TAM directly, or by calling the TSM/TMM.
*/
type TrustAssessmentManager interface {
	SetManagers(managers TafManagers)
	HandleTasInitRequest(cmd command.HandleRequest[tasmsg.TasInitRequest])
	HandleTasTeardownRequest(cmd command.HandleRequest[tasmsg.TasTeardownRequest])
	HandleTasTaRequest(cmd command.HandleRequest[tasmsg.TasTaRequest])
	HandleTasSubscribeRequest(cmd command.HandleSubscriptionRequest[tasmsg.TasSubscribeRequest])
	HandleTasUnsubscribeRequest(cmd command.HandleSubscriptionRequest[tasmsg.TasUnsubscribeRequest])
	DispatchToWorker(session session.Session, tmiID string, cmd core.Command)
	DispatchToWorkerByFullTMIID(fullTMI string, cmd core.Command)
	HandleATLUpdate(cmd command.HandleATLUpdate)
	Sessions() map[string]session.Session
	AddNewTrustModelInstance(instance core.TrustModelInstance, sessionID string)
	RemoveTrustModelInstance(tmiID string, sessionID string)
	QueryTMIs(query string) ([]string, error)
	AddSessionListener(listener listener.SessionListener)
	RemoveSessionListener(listener listener.SessionListener)
	AddATLListener(listener listener.ActualTrustLevelListener)
	RemoveATLListener(listener listener.ActualTrustLevelListener)
	AddTMIListener(listener listener.TrustModelInstanceListener)
	RemoveTMIListener(listener listener.TrustModelInstanceListener)
	Run()
}

/*
The TrustSourceManager is an internal component responsible for handling trust sources, their subscriptions and incoming
evidence messages.
*/
type TrustSourceManager interface {
	SetManagers(managers TafManagers)
	HandleAivResponse(cmd command.HandleResponse[aivmsg.AivResponse])
	HandleAivSubscribeResponse(cmd command.HandleResponse[aivmsg.AivSubscribeResponse])
	HandleAivUnsubscribeResponse(cmd command.HandleResponse[aivmsg.AivUnsubscribeResponse])
	HandleAivNotify(cmd command.HandleNotify[aivmsg.AivNotify])
	HandleMbdSubscribeResponse(cmd command.HandleResponse[mbdmsg.MBDSubscribeResponse])
	HandleMbdUnsubscribeResponse(cmd command.HandleResponse[mbdmsg.MBDUnsubscribeResponse])
	HandleMbdNotify(cmd command.HandleNotify[mbdmsg.MBDNotify])
	HandleTchNotify(cmd command.HandleNotify[tchmsg.TchNotify])
	SubscribeTrustSourceQuantifiers(session session.Session, handler *completionhandler.CompletionHandler)
	UnsubscribeTrustSourceQuantifiers(session session.Session, handler *completionhandler.CompletionHandler)
	RegisterCallback(messageType messages.MessageSchema, requestID string, fn func(cmd core.Command))
	DispatchAivRequest(session session.Session)
}

/*
The TrustMdodelManager is an internal component responsible for handling trust model templates and V2X communication monitoring.
*/
type TrustModelManager interface {
	SetManagers(managers TafManagers)
	HandleV2xCpmMessage(cmd command.HandleOneWay[v2xmsg.V2XCpm])
	ResolveTMT(identifier string) core.TrustModelTemplate
	GetAllTMTs() []core.TrustModelTemplate
	ListRecentV2XNodes() []string
}
