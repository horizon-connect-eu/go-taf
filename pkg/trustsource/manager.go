package trustsource

import (
	"fmt"
	"github.com/google/uuid"
	logging "github.com/vs-uulm/go-taf/internal/logger"
	"github.com/vs-uulm/go-taf/internal/util"
	"github.com/vs-uulm/go-taf/pkg/command"
	"github.com/vs-uulm/go-taf/pkg/communication"
	"github.com/vs-uulm/go-taf/pkg/config"
	"github.com/vs-uulm/go-taf/pkg/core"
	"github.com/vs-uulm/go-taf/pkg/crypto"
	"github.com/vs-uulm/go-taf/pkg/manager"
	messages "github.com/vs-uulm/go-taf/pkg/message"
	aivmsg "github.com/vs-uulm/go-taf/pkg/message/aiv"
	mbdmsg "github.com/vs-uulm/go-taf/pkg/message/mbd"
	tchmsg "github.com/vs-uulm/go-taf/pkg/message/tch"
	"log/slog"
)

type Manager struct {
	config                       config.Configuration
	tafContext                   core.TafContext
	logger                       *slog.Logger
	tam                          manager.TrustAssessmentManager
	tmm                          manager.TrustModelManager
	crypto                       *crypto.Crypto
	outbox                       chan core.Message
	pendingSubscriptionCallbacks map[messages.MessageSchema]map[string]func(cmd core.Command) bool
}

func NewManager(tafContext core.TafContext, channels core.TafChannels) (*Manager, error) {
	tsm := &Manager{
		config:     tafContext.Configuration,
		tafContext: tafContext,
		logger:     logging.CreateChildLogger(tafContext.Logger, "TSM"),
		crypto:     tafContext.Crypto,
		outbox:     channels.OutgoingMessageChannel,
	}

	tsm.pendingSubscriptionCallbacks = make(map[messages.MessageSchema]map[string]func(cmd core.Command) bool)
	tsm.pendingSubscriptionCallbacks[messages.AIV_SUBSCRIBE_RESPONSE] = make(map[string]func(cmd core.Command) bool)
	tsm.pendingSubscriptionCallbacks[messages.AIV_UNSUBSCRIBE_RESPONSE] = make(map[string]func(cmd core.Command) bool)
	tsm.pendingSubscriptionCallbacks[messages.MBD_SUBSCRIBE_RESPONSE] = make(map[string]func(cmd core.Command) bool)
	tsm.pendingSubscriptionCallbacks[messages.MBD_UNSUBSCRIBE_RESPONSE] = make(map[string]func(cmd core.Command) bool)

	tsm.logger.Info("Initializing Trust Source Manager")
	return tsm, nil
}

func (tsm *Manager) SetManagers(managers manager.TafManagers) {
	tsm.tam = managers.TAM
	tsm.tmm = managers.TMM
}

/* ------------ ------------ AIV Message Handling ------------ ------------ */

func (tsm *Manager) HandleAivResponse(cmd command.HandleResponse[aivmsg.AivResponse]) {
	valid, err := tsm.crypto.VerifyAivResponse(&cmd.Response)
	if err != nil {
		tsm.logger.Error("Error verifying AIV_RESPONSE", "Cause", err)
		return
	}
	if !valid {
		tsm.logger.Warn("AIV_RESPONSE could not be verified, ignoring message")
		return
	}
	tsm.logger.Info("TODO: handle AIV_RESPONSE")
}

func (tsm *Manager) HandleAivSubscribeResponse(cmd command.HandleResponse[aivmsg.AivSubscribeResponse]) {
	callback, exists := tsm.pendingSubscriptionCallbacks[messages.AIV_SUBSCRIBE_RESPONSE][cmd.ResponseID]
	if !exists {
		tsm.logger.Warn("AIV_SUBSCRIBE_RESPONSE with unknown response ID received.")
	} else {
		callback(cmd)
		delete(tsm.pendingSubscriptionCallbacks[messages.AIV_SUBSCRIBE_RESPONSE], cmd.ResponseID)
	}
}

func (tsm *Manager) HandleAivUnsubscribeResponse(cmd command.HandleResponse[aivmsg.AivUnsubscribeResponse]) {
	callback, exists := tsm.pendingSubscriptionCallbacks[messages.AIV_UNSUBSCRIBE_RESPONSE][cmd.ResponseID]
	if !exists {
		tsm.logger.Warn("AIV_UNSUBSCRIBE_RESPONSE with unknown response ID received.")
	} else {
		callback(cmd)
		delete(tsm.pendingSubscriptionCallbacks[messages.AIV_UNSUBSCRIBE_RESPONSE], cmd.ResponseID)
	}
}

func (tsm *Manager) HandleAivNotify(cmd command.HandleNotify[aivmsg.AivNotify]) {
	valid, err := tsm.crypto.VerifyAivNotify(&cmd.Notify)
	if err != nil {
		tsm.logger.Error("Error verifying AIV_NOTIFY", "Cause", err)
		return
	}
	if !valid {
		tsm.logger.Warn("AIV_NOTIFY could not be verified, ignoring message")
		return
	}
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

func (tsm *Manager) InitTrustSourceQuantifiers(tmi core.TrustModelInstance) map[string]func(cmd core.Command) bool {

	callbacks := make(map[string]func(cmd core.Command) bool, 0)

	subscriptions := make(map[core.Source]map[string][]core.Evidence, 0)

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
				AttestationCertificate: tsm.crypto.AttestationCertificate(),
				CheckInterval:          1000,
				Evidence:               aivmsg.AIVSUBSCRIBEREQUESTEvidence{},
				Subscribe:              subscribeField,
			}
			tsm.crypto.SignAivSubscribeRequest(&subMsg)
			subReqId := tsm.GenerateRequestId()
			bytes, err := communication.BuildSubscriptionRequest(tsm.config.Communication.TafEndpoint, messages.AIV_SUBSCRIBE_REQUEST, tsm.config.Communication.TafEndpoint, tsm.config.Communication.TafEndpoint, subReqId, subMsg)
			if err != nil {
				tsm.logger.Error("Error marshalling response", "error", err)
			}

			callbacks[subReqId] = func(recvCmd core.Command) bool {
				switch cmd := recvCmd.(type) {
				case command.HandleResponse[aivmsg.AivSubscribeResponse]:
					tsm.logger.Warn(*cmd.Response.SubscriptionID)
					return true
				default:
					return false
				}
				return true
			}
			tsm.RegisterCallback(messages.AIV_SUBSCRIBE_RESPONSE, subReqId, callbacks[subReqId])
			//TODO: What does InitTrustSourceQuantifiers return instead?

			//Send response message
			tsm.outbox <- core.NewMessage(bytes, "", tsm.config.Communication.AivEndpoint)
			util.UNUSED(subReqId)

		case core.MBD:
			subMsg := mbdmsg.MBDSubscribeRequest{
				AttestationCertificate: tsm.crypto.AttestationCertificate(),
				Subscribe:              true,
			}
			subReqId := tsm.GenerateRequestId()
			bytes, err := communication.BuildSubscriptionRequest(tsm.config.Communication.TafEndpoint, messages.MBD_SUBSCRIBE_REQUEST, tsm.config.Communication.TafEndpoint, tsm.config.Communication.TafEndpoint, subReqId, subMsg)
			if err != nil {
				tsm.logger.Error("Error marshalling response", "error", err)
			}
			//Send response message
			tsm.outbox <- core.NewMessage(bytes, "", tsm.config.Communication.MbdEndpoint)
			util.UNUSED(subReqId)
		default:
			panic("unknown Trust Source")
		}
	}
	return callbacks
}

func (tsm *Manager) RegisterCallback(messageType messages.MessageSchema, requestID string, fn func(cmd core.Command) bool) {
	tsm.pendingSubscriptionCallbacks[messageType][requestID] = fn
}

func (tsm *Manager) GenerateRequestId() string {
	//When debug configuration provides fixed session ID, use this ID
	if tsm.config.Debug.FixedRequestID != "" {
		return tsm.config.Debug.FixedRequestID
	} else {
		return "REQ-" + uuid.New().String()
	}
}
