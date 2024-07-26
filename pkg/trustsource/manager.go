package trustsource

import (
	"fmt"
	logging "github.com/vs-uulm/go-taf/internal/logger"
	"github.com/vs-uulm/go-taf/internal/util"
	"github.com/vs-uulm/go-taf/pkg/command"
	"github.com/vs-uulm/go-taf/pkg/communication"
	"github.com/vs-uulm/go-taf/pkg/core"
	"github.com/vs-uulm/go-taf/pkg/manager"
	messages "github.com/vs-uulm/go-taf/pkg/message"
	aivmsg "github.com/vs-uulm/go-taf/pkg/message/aiv"
	mbdmsg "github.com/vs-uulm/go-taf/pkg/message/mbd"
	tchmsg "github.com/vs-uulm/go-taf/pkg/message/tch"
	"log/slog"
)

type Manager struct {
	tafContext core.TafContext
	channels   core.TafChannels
	logger     *slog.Logger
	tam        manager.TrustAssessmentManager
	tmm        manager.TrustModelManager
}

func NewManager(tafContext core.TafContext, channels core.TafChannels) (*Manager, error) {
	tsm := &Manager{
		tafContext: tafContext,
		channels:   channels,
		logger:     logging.CreateChildLogger(tafContext.Logger, "TSM"),
	}
	tsm.logger.Info("Initializing Trust Source Manager")
	return tsm, nil
}

func (tsm *Manager) SetManagers(managers manager.TafManagers) {
	tsm.tam = managers.TAM
	tsm.tmm = managers.TMM
}

/* ------------ ------------ AIV Message Handling ------------ ------------ */

func (tsm *Manager) HandleAivResponse(cmd command.HandleResponse[aivmsg.AivResponse]) {
	tsm.logger.Info("TODO: handle AIV_RESPONSE")
}

func (tsm *Manager) HandleAivSubscribeResponse(cmd command.HandleResponse[aivmsg.AivSubscribeResponse]) {
	tsm.logger.Info("TODO: handle AIV_SUBSCRIBE_RESPONSE")
}

func (tsm *Manager) HandleAivUnsubscribeResponse(cmd command.HandleResponse[aivmsg.AivUnsubscribeResponse]) {
	tsm.logger.Info("TODO: handle AIV_UNSUBSCRIBE_RESPONSE")
}

func (tsm *Manager) HandleAivNotify(cmd command.HandleNotify[aivmsg.AivNotify]) {
	tsm.logger.Info("TODO: handle AIV_NOTIFY")
}

/* ------------ ------------ MBD Message Handling ------------ ------------ */

func (tsm *Manager) HandleMbdSubscribeResponse(cmd command.HandleResponse[mbdmsg.MBDSubscribeResponse]) {
	tsm.logger.Info("TODO: handle MBD_SUBSCRIBE_RESPONSE")
}

func (tsm *Manager) HandleMbdUnsubscribeResponse(cmd command.HandleResponse[mbdmsg.MBDUnsubscribeResponse]) {
	tsm.logger.Info("TODO: handle MBD_UNSUBSCRIBE_RESPONSE")
}

func (tsm *Manager) HandleMbdNotify(cmd command.HandleNotify[mbdmsg.MBDNotify]) {
	tsm.logger.Info("TODO: handle MBD_NOTIFY")
}

func (tsm *Manager) HandleTchNotify(cmd command.HandleNotify[tchmsg.Message]) {

}

func (tsm *Manager) InitTrustSourceQuantifiers(tmi core.TrustModelInstance) {

	subscriptions := make(map[core.Source]map[string][]core.Evidence, 1)

	for _, quantifier := range tmi.TrustSourceQuantifiers() {

		for _, evidence := range quantifier.Evidence {
			if subscriptions[evidence.Source()] == nil {
				subscriptions[evidence.Source()] = make(map[string][]core.Evidence, 0)
			}
			if subscriptions[evidence.Source()][quantifier.Trustee] == nil {
				subscriptions[evidence.Source()][quantifier.Trustee] = make([]core.Evidence, 0)
			}
			subscriptions[evidence.Source()][quantifier.Trustee] = append(subscriptions[evidence.Source()][quantifier.Trustee], evidence)
		}
		//	trustSource.quantifier.Evidence[0].Source()
	}

	fmt.Printf("%+v", subscriptions)

	for source, trustees := range subscriptions {
		switch source {
		case core.AIV:

			subscribeField := make([]aivmsg.Subscribe, 0)

			for trusteeID, evidenceList := range trustees {

				evidenceStringList := make([]string, 0)
				for _, evidence := range evidenceList {
					evidenceStringList = append(evidenceStringList, evidence.String())
				}

				subscribeField = append(subscribeField, aivmsg.Subscribe{
					TrusteeID:       trusteeID,
					RequestedClaims: evidenceStringList,
				})
			}

			subMsg := aivmsg.AivSubscribeRequest{
				AttestationCertificate: "",
				CheckInterval:          1000,
				Evidence:               aivmsg.AIVSUBSCRIBEREQUESTEvidence{},
				Subscribe:              subscribeField,
			}
			bytes, subReqId, err := communication.BuildSubscriptionRequest("taf", messages.AIV_SUBSCRIBE_REQUEST, "taf", "taf", subMsg)
			if err != nil {
				tsm.logger.Error("Error marshalling response", "error", err)
			}
			//Send response message
			tsm.channels.OutgoingMessageChannel <- core.NewMessage(bytes, "", "aiv")
			util.UNUSED(subReqId)

		case core.MBD:
			subMsg := mbdmsg.MBDSubscribeRequest{
				AttestationCertificate: "",
				Subscribe:              true,
			}
			bytes, subReqId, err := communication.BuildSubscriptionRequest("taf", messages.MBD_SUBSCRIBE_REQUEST, "taf", "taf", subMsg)
			if err != nil {
				tsm.logger.Error("Error marshalling response", "error", err)
			}
			//Send response message
			tsm.channels.OutgoingMessageChannel <- core.NewMessage(bytes, "", "mbd")
			util.UNUSED(subReqId)
		default:
			panic("unknown Trust Source")
		}
	}

}
