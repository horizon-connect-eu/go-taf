package trustsource

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/vs-uulm/go-subjectivelogic/pkg/subjectivelogic"
	"github.com/vs-uulm/go-taf/internal/flow/completionhandler"
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
	"github.com/vs-uulm/go-taf/pkg/trustmodel/trustmodelupdate"
	"log/slog"
	"math"
)

const MISSING_EVIDENCE = math.MinInt

type Manager struct {
	config                  config.Configuration
	tafContext              core.TafContext
	logger                  *slog.Logger
	tam                     manager.TrustAssessmentManager
	tmm                     manager.TrustModelManager
	crypto                  *crypto.Crypto
	outbox                  chan core.Message
	pendingMessageCallbacks map[messages.MessageSchema]map[string]func(cmd core.Command)
	subscriptionIDtoTMI     map[string]string
	//subscriptionID:Trustee:Source:EvidenceType->Value
	subscriptionEvidence map[string]map[string]map[core.TrustSource]map[core.EvidenceType]int
	//subscriptionID:Trustee:Source->QuantifierFunc
	subscriptionQuantifiers map[string]map[string]map[core.TrustSource]func(map[core.EvidenceType]int) subjectivelogic.QueryableOpinion
}

func NewManager(tafContext core.TafContext, channels core.TafChannels) (*Manager, error) {
	tsm := &Manager{
		config:                  tafContext.Configuration,
		tafContext:              tafContext,
		logger:                  logging.CreateChildLogger(tafContext.Logger, "TSM"),
		crypto:                  tafContext.Crypto,
		outbox:                  channels.OutgoingMessageChannel,
		subscriptionIDtoTMI:     make(map[string]string),
		subscriptionEvidence:    make(map[string]map[string]map[core.TrustSource]map[core.EvidenceType]int),
		subscriptionQuantifiers: make(map[string]map[string]map[core.TrustSource]func(map[core.EvidenceType]int) subjectivelogic.QueryableOpinion),
	}

	tsm.pendingMessageCallbacks = map[messages.MessageSchema]map[string]func(cmd core.Command){
		messages.AIV_SUBSCRIBE_RESPONSE:   make(map[string]func(cmd core.Command)),
		messages.AIV_UNSUBSCRIBE_RESPONSE: make(map[string]func(cmd core.Command)),
		messages.MBD_SUBSCRIBE_RESPONSE:   make(map[string]func(cmd core.Command)),
		messages.MBD_UNSUBSCRIBE_RESPONSE: make(map[string]func(cmd core.Command)),
		messages.AIV_RESPONSE:             make(map[string]func(cmd core.Command)),
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
	valid, err := tsm.crypto.VerifyAivResponse(&cmd.Response)
	if err != nil {
		tsm.logger.Error("Error verifying AIV_RESPONSE", "Cause", err)
		return
	}
	if !valid {
		tsm.logger.Warn("AIV_RESPONSE could not be verified, ignoring message")
		return
	}
	callback, exists := tsm.pendingMessageCallbacks[messages.AIV_RESPONSE][cmd.ResponseID]
	if !exists {
		tsm.logger.Warn("AIV_RESPONSE with unknown response ID received.")
	} else {
		callback(cmd)
		delete(tsm.pendingMessageCallbacks[messages.AIV_RESPONSE], cmd.ResponseID)
	}
}

func (tsm *Manager) HandleAivSubscribeResponse(cmd command.HandleResponse[aivmsg.AivSubscribeResponse]) {
	callback, exists := tsm.pendingMessageCallbacks[messages.AIV_SUBSCRIBE_RESPONSE][cmd.ResponseID]
	if !exists {
		tsm.logger.Warn("AIV_SUBSCRIBE_RESPONSE with unknown response ID received.")
	} else {
		callback(cmd)
		delete(tsm.pendingMessageCallbacks[messages.AIV_SUBSCRIBE_RESPONSE], cmd.ResponseID)
	}
}

func (tsm *Manager) HandleAivUnsubscribeResponse(cmd command.HandleResponse[aivmsg.AivUnsubscribeResponse]) {
	callback, exists := tsm.pendingMessageCallbacks[messages.AIV_UNSUBSCRIBE_RESPONSE][cmd.ResponseID]
	if !exists {
		tsm.logger.Warn("AIV_UNSUBSCRIBE_RESPONSE with unknown response ID received.")
	} else {
		callback(cmd)
		delete(tsm.pendingMessageCallbacks[messages.AIV_UNSUBSCRIBE_RESPONSE], cmd.ResponseID)
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

	subscriptionID := *cmd.Notify.SubscriptionID
	tmiID, exists := tsm.subscriptionIDtoTMI[subscriptionID]
	util.UNUSED(tmiID)
	if !exists {
		return
	}

	for _, trusteeReport := range cmd.Notify.TrusteeReports {
		for _, report := range trusteeReport.AttestationReport {
			evidence := report.Claim
			tsm.logger.Debug(fmt.Sprintf("%+v", tsm.subscriptionEvidence))
			tsm.logger.Debug("sub Id: " + subscriptionID)
			tsm.logger.Debug("Trustee Id: " + *trusteeReport.TrusteeID)
			tsm.logger.Debug("Source: " + string(core.AIV))
			tsm.logger.Debug("Evidence: " + core.EvidenceTypeByName(evidence).String())
			tsm.subscriptionEvidence[subscriptionID][*trusteeReport.TrusteeID][core.AIV][core.EvidenceTypeByName(evidence)] = int(report.Appraisal)
		}
		evidence := tsm.subscriptionEvidence[subscriptionID][*trusteeReport.TrusteeID][core.AIV]
		//call quantifier
		ato := tsm.subscriptionQuantifiers[subscriptionID][*trusteeReport.TrusteeID][core.AIV](evidence)
		//create update operation
		update := trustmodelupdate.CreateAtomicTrustOpinionUpdate(ato, *trusteeReport.TrusteeID, core.AIV)
		tmiUpdateCmd := command.CreateHandleTMIUpdate(tmiID, update)
		util.UNUSED(tmiUpdateCmd)
		tsm.logger.Warn("Opinion for " + *trusteeReport.TrusteeID + ": " + ato.String())
	}

}

/* ------------ ------------ MBD Message Handling ------------ ------------ */

func (tsm *Manager) HandleMbdSubscribeResponse(cmd command.HandleResponse[mbdmsg.MBDSubscribeResponse]) {
	callback, exists := tsm.pendingMessageCallbacks[messages.MBD_SUBSCRIBE_RESPONSE][cmd.ResponseID]
	if !exists {
		tsm.logger.Warn("MBD_SUBSCRIBE_RESPONSE with unknown response ID received.")
	} else {
		callback(cmd)
		delete(tsm.pendingMessageCallbacks[messages.MBD_SUBSCRIBE_RESPONSE], cmd.ResponseID)
	}
}

func (tsm *Manager) HandleMbdUnsubscribeResponse(cmd command.HandleResponse[mbdmsg.MBDUnsubscribeResponse]) {
	callback, exists := tsm.pendingMessageCallbacks[messages.MBD_UNSUBSCRIBE_RESPONSE][cmd.ResponseID]
	if !exists {
		tsm.logger.Warn("MBD_UNSUBSCRIBE_RESPONSE with unknown response ID received.")
	} else {
		callback(cmd)
		delete(tsm.pendingMessageCallbacks[messages.MBD_UNSUBSCRIBE_RESPONSE], cmd.ResponseID)
	}
}

func (tsm *Manager) HandleMbdNotify(cmd command.HandleNotify[mbdmsg.MBDNotify]) {
	tsm.logger.Info("TODO: handle MBD_NOTIFY")
}

func (tsm *Manager) HandleTchNotify(cmd command.HandleNotify[tchmsg.Message]) {

}

func (tsm *Manager) InitializeTrustSourceQuantifiers(tmt core.TrustModelTemplate, trustModelInstanceID string, handler *completionhandler.CompletionHandler) {

	subscriptions := make(map[core.TrustSource]map[string][]core.EvidenceType, 0)
	quantifiers := make(map[core.TrustSource]core.Quantifier, 0)

	for _, item := range tmt.TrustSourceQuantifiers() {

		quantifiers[item.TrustSource] = item.Quantifier
		for _, evidence := range item.Evidence {
			if subscriptions[evidence.Source()] == nil {
				subscriptions[evidence.Source()] = make(map[string][]core.EvidenceType, 0)
			}
			if subscriptions[evidence.Source()][item.Trustee] == nil {
				subscriptions[evidence.Source()][item.Trustee] = make([]core.EvidenceType, 0)
			}
			subscriptions[evidence.Source()][item.Trustee] = append(subscriptions[evidence.Source()][item.Trustee], evidence)

		}
		//	trustSource.quantifier.Evidence[0].Source()
	}

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

			resolve, reject := handler.Register()
			tsm.RegisterCallback(messages.AIV_SUBSCRIBE_RESPONSE, subReqId, func(recvCmd core.Command) {
				switch cmd := recvCmd.(type) {
				case command.HandleResponse[aivmsg.AivSubscribeResponse]:
					if cmd.Response.Error != nil {
						reject(errors.New(*cmd.Response.Error))
						return
					}
					tsm.subscriptionIDtoTMI[*cmd.Response.SubscriptionID] = trustModelInstanceID
					tsm.subscriptionEvidence[*cmd.Response.SubscriptionID] = make(map[string]map[core.TrustSource]map[core.EvidenceType]int)
					tsm.subscriptionQuantifiers[*cmd.Response.SubscriptionID] = make(map[string]map[core.TrustSource]func(map[core.EvidenceType]int) subjectivelogic.QueryableOpinion)

					for trusteeID, evidenceList := range trustees {
						tsm.subscriptionEvidence[*cmd.Response.SubscriptionID][trusteeID] = make(map[core.TrustSource]map[core.EvidenceType]int)
						tsm.subscriptionEvidence[*cmd.Response.SubscriptionID][trusteeID][core.AIV] = make(map[core.EvidenceType]int)
						tsm.subscriptionQuantifiers[*cmd.Response.SubscriptionID][trusteeID] = make(map[core.TrustSource]func(map[core.EvidenceType]int) subjectivelogic.QueryableOpinion)
						tsm.subscriptionQuantifiers[*cmd.Response.SubscriptionID][trusteeID][core.AIV] = quantifiers[core.AIV]
						for _, evidence := range evidenceList {
							tsm.subscriptionEvidence[*cmd.Response.SubscriptionID][trusteeID][core.AIV][evidence] = MISSING_EVIDENCE
						}
					}
					resolve()
				default:
					reject(errors.New("Unknown response type: " + cmd.Type().String()))
				}
			})
			//Send response message
			tsm.outbox <- core.NewMessage(bytes, "", tsm.config.Communication.AivEndpoint)

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

			resolve, reject := handler.Register()
			tsm.RegisterCallback(messages.MBD_SUBSCRIBE_RESPONSE, subReqId, func(recvCmd core.Command) {
				switch cmd := recvCmd.(type) {
				case command.HandleResponse[mbdmsg.MBDSubscribeResponse]:
					if cmd.Response.Error != nil {
						reject(errors.New(*cmd.Response.Error))
						//TODO: remove session
						return
					}
					tsm.logger.Warn(*cmd.Response.SubscriptionID)
					resolve()
				default:
					reject(errors.New("Unknown response type: " + cmd.Type().String()))
				}
			})
			//Send response message
			tsm.outbox <- core.NewMessage(bytes, "", tsm.config.Communication.MbdEndpoint)
			util.UNUSED(subReqId)
		case core.TCH:
			//TODO
		default:
			panic("unknown Trust Source")
		}
	}

}

/*
The RegisterCallback function adds a callback for a given Message Type and Request ID (== expected Response ID)
*/
func (tsm *Manager) RegisterCallback(messageType messages.MessageSchema, requestID string, fn func(cmd core.Command)) {
	tsm.pendingMessageCallbacks[messageType][requestID] = fn
}

func (tsm *Manager) GenerateRequestId() string {
	//When debug configuration provides fixed session ID, use this ID
	if tsm.config.Debug.FixedRequestID != "" {
		return tsm.config.Debug.FixedRequestID
	} else {
		return "REQ-" + uuid.New().String()
	}
}
