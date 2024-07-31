package manager

import (
	"github.com/vs-uulm/go-taf/internal/flow/completionhandler"
	"github.com/vs-uulm/go-taf/pkg/command"
	"github.com/vs-uulm/go-taf/pkg/core"
	messages "github.com/vs-uulm/go-taf/pkg/message"
	aivmsg "github.com/vs-uulm/go-taf/pkg/message/aiv"
	mbdmsg "github.com/vs-uulm/go-taf/pkg/message/mbd"
	tasmsg "github.com/vs-uulm/go-taf/pkg/message/tas"
	tchmsg "github.com/vs-uulm/go-taf/pkg/message/tch"
	v2xmsg "github.com/vs-uulm/go-taf/pkg/message/v2x"
)

type TafManagers struct {
	TSM TrustSourceManager
	TAM TrustAssessmentManager
	TMM TrustModelManager
}

/*
The CooperativeManager is a manager type that knows about other managers (via SetManagers) and can thus call their functions .
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

type TrustAssessmentManager interface {
	SetManagers(managers TafManagers)
	HandleTasInitRequest(cmd command.HandleRequest[tasmsg.TasInitRequest])
	HandleTasTeardownRequest(cmd command.HandleRequest[tasmsg.TasTeardownRequest])
	HandleTasTaRequest(cmd command.HandleRequest[tasmsg.TasTaRequest])
	HandleTasSubscribeRequest(cmd command.HandleSubscriptionRequest[tasmsg.TasSubscribeRequest])
	HandleTasUnsubscribeRequest(cmd command.HandleSubscriptionRequest[tasmsg.TasUnsubscribeRequest])
	DispatchToWorker(tmiID string, cmd core.Command)
	HandleATLUpdate(cmd command.HandleATLUpdate)
	Run()
}

type TrustSourceManager interface {
	SetManagers(managers TafManagers)
	HandleAivResponse(cmd command.HandleResponse[aivmsg.AivResponse])
	HandleAivSubscribeResponse(cmd command.HandleResponse[aivmsg.AivSubscribeResponse])
	HandleAivUnsubscribeResponse(cmd command.HandleResponse[aivmsg.AivUnsubscribeResponse])
	HandleAivNotify(cmd command.HandleNotify[aivmsg.AivNotify])
	HandleMbdSubscribeResponse(cmd command.HandleResponse[mbdmsg.MBDSubscribeResponse])
	HandleMbdUnsubscribeResponse(cmd command.HandleResponse[mbdmsg.MBDUnsubscribeResponse])
	HandleMbdNotify(cmd command.HandleNotify[mbdmsg.MBDNotify])
	HandleTchNotify(cmd command.HandleNotify[tchmsg.Message])
	InitializeTrustSourceQuantifiers(tmt core.TrustModelTemplate, trustModelInstanceID string, handler *completionhandler.CompletionHandler)
	GenerateRequestId() string
	RegisterCallback(messageType messages.MessageSchema, requestID string, fn func(cmd core.Command))
}

type TrustModelManager interface {
	SetManagers(managers TafManagers)
	HandleV2xCpmMessage(cmd command.HandleOneWay[v2xmsg.V2XCpm])
	ResolveTMT(identifier string) core.TrustModelTemplate
}
