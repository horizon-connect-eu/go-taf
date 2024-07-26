package manager

import (
	"github.com/vs-uulm/go-taf/pkg/command"
	"github.com/vs-uulm/go-taf/pkg/core"
	aivmsg "github.com/vs-uulm/go-taf/pkg/message/aiv"
	mbdmsg "github.com/vs-uulm/go-taf/pkg/message/mbd"
	tasmsg "github.com/vs-uulm/go-taf/pkg/message/tas"
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

type TrustSourceManager interface {
	SetManagers(managers TafManagers)
	HandleAivResponse(cmd command.HandleResponse[aivmsg.AivResponse])
	HandleAivSubscribeResponse(cmd command.HandleResponse[aivmsg.AivSubscribeResponse])
	HandleAivUnsubscribeResponse(cmd command.HandleResponse[aivmsg.AivUnsubscribeResponse])
	HandleAivNotify(cmd command.HandleNotify[aivmsg.AivNotify])
	HandleMbdSubscribeResponse(cmd command.HandleResponse[mbdmsg.MBDSubscribeResponse])
	HandleMbdUnsubscribeResponse(cmd command.HandleResponse[mbdmsg.MBDUnsubscribeResponse])
	HandleMbdNotify(cmd command.HandleNotify[mbdmsg.MBDNotify])
	InitTrustSourceQuantifiers(tmi core.TrustModelInstance)
}

type TrustAssessmentManager interface {
	SetManagers(managers TafManagers)
	HandleTasInitRequest(cmd command.HandleRequest[tasmsg.TasInitRequest])
	HandleTasTeardownRequest(cmd command.HandleRequest[tasmsg.TasTeardownRequest])
	Run()
}

type TrustModelManager interface {
	SetManagers(managers TafManagers)
	HandleV2xCpmMessage(cmd command.HandleOneWay[v2xmsg.V2XCpm])
	ResolveTMT(identifier string) core.TrustModelTemplate
}
