package trustsource

import (
	"errors"
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
	"github.com/vs-uulm/go-taf/pkg/trustmodel/session"
	"github.com/vs-uulm/go-taf/pkg/trustmodel/trustmodelupdate"
	"log/slog"
	"math"
)

const MISSING_EVIDENCE = math.MinInt

type Manager struct {
	config     config.Configuration
	tafContext core.TafContext
	logger     *slog.Logger
	tam        manager.TrustAssessmentManager
	tmm        manager.TrustModelManager
	crypto     *crypto.Crypto
	outbox     chan core.Message
	//Schema:ResponseID->Callback
	pendingMessageCallbacks map[messages.MessageSchema]map[string]func(cmd core.Command)
	//subscriptionID->TMI ID
	subscriptionIDtoTMI map[string]string
	//TMI ID:subscriptionID->bool
	tmiToSubscriptionID map[string]map[string]core.TrustSource
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
		tmiToSubscriptionID:     make(map[string]map[string]core.TrustSource),
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

	updates := make([]core.Update, 0)

	for _, trusteeReport := range cmd.Notify.TrusteeReports {
		for _, report := range trusteeReport.AttestationReport {
			evidence := report.Claim
			tsm.logger.Debug("Received new evidence from AIV", "SubscriptionID", subscriptionID, "Source", core.AIV.String(), "Evidence Type", core.EvidenceTypeByName(evidence).String(), "Trustee ID", *trusteeReport.TrusteeID)
			tsm.subscriptionEvidence[subscriptionID][*trusteeReport.TrusteeID][core.AIV][core.EvidenceTypeByName(evidence)] = int(report.Appraisal)
		}
		evidence := tsm.subscriptionEvidence[subscriptionID][*trusteeReport.TrusteeID][core.AIV]
		//call quantifier
		ato := tsm.subscriptionQuantifiers[subscriptionID][*trusteeReport.TrusteeID][core.AIV](evidence)
		tsm.logger.Info("Opinion for " + *trusteeReport.TrusteeID + ": " + ato.String())
		//create update operation
		update := trustmodelupdate.CreateAtomicTrustOpinionUpdate(ato, *trusteeReport.TrusteeID, core.AIV)
		updates = append(updates, update)
	}
	if len(updates) > 0 {
		tmiUpdateCmd := command.CreateHandleTMIUpdate(tmiID, updates...)
		tsm.tam.DispatchToWorker(tmiID, tmiUpdateCmd)
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
	util.UNUSED(cmd)
}

func (tsm *Manager) HandleTchNotify(cmd command.HandleNotify[tchmsg.Message]) {
	tsm.logger.Info("TODO: handle TCH_NOTIFY")
	util.UNUSED(cmd)
}

func (tsm *Manager) SubscribeTrustSourceQuantifiers(tmt core.TrustModelTemplate, trustModelInstanceID string, handler *completionhandler.CompletionHandler) {

	//When no handler has been set, create empty one
	if handler == nil {
		handler = completionhandler.New(func() {}, func(err error) {
		})
		defer handler.Execute()
	}

	subscriptions := make(map[core.TrustSource]map[string][]core.EvidenceType)
	quantifiers := make(map[core.TrustSource]core.Quantifier)

	for _, item := range tmt.TrustSourceQuantifiers() {

		quantifiers[item.TrustSource] = item.Quantifier
		for _, evidence := range item.Evidence {
			if subscriptions[evidence.Source()] == nil {
				subscriptions[evidence.Source()] = make(map[string][]core.EvidenceType)
			}
			if subscriptions[evidence.Source()][item.Trustee] == nil {
				subscriptions[evidence.Source()][item.Trustee] = make([]core.EvidenceType, 0)
			}
			subscriptions[evidence.Source()][item.Trustee] = append(subscriptions[evidence.Source()][item.Trustee], evidence)

		}
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
				CheckInterval:          int64(tsm.config.Evidence.AIV.CheckInterval),
				Evidence:               aivmsg.AIVSUBSCRIBEREQUESTEvidence{},
				Subscribe:              subscribeField,
			}
			err := tsm.crypto.SignAivSubscribeRequest(&subMsg)
			util.UNUSED(err)
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

					//add to map: TMI=>sub ID
					_, exists := tsm.tmiToSubscriptionID[trustModelInstanceID]
					if !exists {
						tsm.tmiToSubscriptionID[trustModelInstanceID] = make(map[string]core.TrustSource)
					}
					tsm.tmiToSubscriptionID[trustModelInstanceID][*cmd.Response.SubscriptionID] = core.AIV

					//add to map: sub ID=>TMI
					tsm.subscriptionIDtoTMI[*cmd.Response.SubscriptionID] = trustModelInstanceID

					//create deeper maps
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
						//TODO: remove session for MBD
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
			//Nothing to do here (yet)
		default:
			panic("unknown Trust Source")
		}
	}
}

func (tsm *Manager) UnsubscribeTrustSourceQuantifiers(tmt core.TrustModelTemplate, trustModelInstanceID string, handler *completionhandler.CompletionHandler) {
	util.UNUSED(tmt)

	//When no handler has been set, create empty one
	if handler == nil {
		handler = completionhandler.New(func() {}, func(err error) {
		})
		defer handler.Execute()
	}

	//Get Subscription ID(s) for TMI ID
	subIDs, exists := tsm.tmiToSubscriptionID[trustModelInstanceID]
	if !exists {
		_, reject := handler.Register()
		reject(errors.New("Unknown trust model instance ID: " + trustModelInstanceID))
		return
	}
	for subID, source := range subIDs {

		resolve, reject := handler.Register()

		switch source {
		case core.AIV:
			unsubMsg := aivmsg.AivUnsubscribeRequest{
				AttestationCertificate: tsm.crypto.AttestationCertificate(),
				SubscriptionID:         subID,
			}
			unsubReqId := tsm.GenerateRequestId()
			bytes, err := communication.BuildSubscriptionRequest(tsm.config.Communication.TafEndpoint, messages.AIV_UNSUBSCRIBE_REQUEST, tsm.config.Communication.TafEndpoint, tsm.config.Communication.TafEndpoint, unsubReqId, unsubMsg)
			if err != nil {
				tsm.logger.Error("Error marshalling response", "error", err)
			}
			tsm.RegisterCallback(messages.AIV_UNSUBSCRIBE_RESPONSE, unsubReqId, func(recvCmd core.Command) {
				switch cmd := recvCmd.(type) {
				case command.HandleResponse[aivmsg.AivUnsubscribeResponse]:
					if cmd.Response.Error != nil {
						reject(errors.New(*cmd.Response.Error))
						return
					}
					//delete associated data structures/lookups
					delete(tsm.subscriptionIDtoTMI, subID)
					delete(tsm.subscriptionEvidence, subID)
					delete(tsm.subscriptionQuantifiers, subID)
					delete(tsm.tmiToSubscriptionID[trustModelInstanceID], subID)

					tsm.logger.Debug("Unregistering Subscription " + subID)

					resolve()
				default:
					reject(errors.New("Unknown response type: " + cmd.Type().String()))
				}
			})
			tsm.outbox <- core.NewMessage(bytes, "", tsm.config.Communication.AivEndpoint)
		case core.MBD:
			unsubMsg := mbdmsg.MBDUnsubscribeRequest{
				AttestationCertificate: tsm.crypto.AttestationCertificate(),
				SubscriptionID:         subID,
			}
			unsubReqId := tsm.GenerateRequestId()
			bytes, err := communication.BuildSubscriptionRequest(tsm.config.Communication.TafEndpoint, messages.MBD_UNSUBSCRIBE_REQUEST, tsm.config.Communication.TafEndpoint, tsm.config.Communication.TafEndpoint, unsubReqId, unsubMsg)
			if err != nil {
				tsm.logger.Error("Error marshalling response", "error", err)
			}
			tsm.RegisterCallback(messages.MBD_UNSUBSCRIBE_REQUEST, unsubReqId, func(recvCmd core.Command) {
				switch cmd := recvCmd.(type) {
				case command.HandleResponse[aivmsg.AivUnsubscribeResponse]:
					if cmd.Response.Error != nil {
						reject(errors.New(*cmd.Response.Error))
						return
					}
					//delete associated data structures/lookups
					delete(tsm.subscriptionIDtoTMI, subID)
					delete(tsm.subscriptionEvidence, subID)
					delete(tsm.subscriptionQuantifiers, subID)
					delete(tsm.tmiToSubscriptionID[trustModelInstanceID], subID)

					tsm.logger.Debug("Unregistering Subscription " + subID)

					resolve()
				default:
					reject(errors.New("Unknown response type: " + cmd.Type().String()))
				}
			})
			tsm.outbox <- core.NewMessage(bytes, "", tsm.config.Communication.MbdEndpoint)
		case core.TCH:
			//Nothing to do here (yet)
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

func (tsm *Manager) DispatchAivRequest(session session.Session) {

	tmt := session.TrustModelTemplate()

	query := make(map[core.TrustSource]map[string][]core.EvidenceType)
	quantifiers := make(map[core.TrustSource]core.Quantifier)

	for _, item := range tmt.TrustSourceQuantifiers() {

		quantifiers[item.TrustSource] = item.Quantifier
		for _, evidence := range item.Evidence {
			if query[evidence.Source()] == nil {
				query[evidence.Source()] = make(map[string][]core.EvidenceType)
			}
			if query[evidence.Source()][item.Trustee] == nil {
				query[evidence.Source()][item.Trustee] = make([]core.EvidenceType, 0)
			}
			query[evidence.Source()][item.Trustee] = append(query[evidence.Source()][item.Trustee], evidence)

		}
	}

	trustees, exists := query[core.AIV]
	if exists {

		queryField := make([]aivmsg.Query, 0)

		for trusteeID, evidenceList := range trustees {

			evidenceStringList := make([]string, 0)
			for _, evidence := range evidenceList {
				evidenceStringList = append(evidenceStringList, evidence.String())
			}

			queryField = append(queryField, aivmsg.Query{
				TrusteeID:       trusteeID,
				RequestedClaims: evidenceStringList,
			})
		}
		reqMsg := aivmsg.AivRequest{
			AttestationCertificate: tsm.crypto.AttestationCertificate(),
			Evidence:               aivmsg.AIVREQUESTEvidence{},
			Query:                  queryField,
		}
		err := tsm.crypto.SignAivRequest(&reqMsg)
		util.UNUSED(err)

		reqId := tsm.GenerateRequestId()
		bytes, err := communication.BuildRequest(tsm.config.Communication.TafEndpoint, messages.AIV_REQUEST, tsm.config.Communication.TafEndpoint, reqId, reqMsg)
		if err != nil {
			tsm.logger.Error("Error marshalling request", "error", err)
		}

		tsm.RegisterCallback(messages.AIV_RESPONSE, reqId, func(recvCmd core.Command) {
			switch cmd := recvCmd.(type) {
			case command.HandleResponse[aivmsg.AivResponse]:
				for _, trusteeReport := range cmd.Response.TrusteeReports {
					evidenceCollection := make(map[core.EvidenceType]int)
					for _, report := range trusteeReport.AttestationReport {
						evidence := report.Claim
						tsm.logger.Debug("Received evidence response from AIV", "Evidence Type", core.EvidenceTypeByName(evidence).String(), "Trustee ID", *trusteeReport.TrusteeID)
						evidenceCollection[core.EvidenceTypeByName(evidence)] = int(report.Appraisal)
					}
					//call quantifier
					ato := quantifiers[core.AIV](evidenceCollection)
					tsm.logger.Info("Opinion for " + *trusteeReport.TrusteeID + ": " + ato.String())
					//create update operation for all TMIs of session
					for tmiID := range session.TrustModelInstances() {
						update := trustmodelupdate.CreateAtomicTrustOpinionUpdate(ato, *trusteeReport.TrusteeID, core.AIV)
						tmiUpdateCmd := command.CreateHandleTMIUpdate(tmiID, update)
						tsm.tam.DispatchToWorker(tmiID, tmiUpdateCmd)
					}
				}
			default:
				//Nothing to do
			}
		})
		//Send response message
		tsm.outbox <- core.NewMessage(bytes, "", tsm.config.Communication.AivEndpoint)
	}
}

func (tsm *Manager) GenerateRequestId() string {
	//When debug configuration provides fixed session ID, use this ID
	if tsm.config.Debug.FixedRequestID != "" {
		return tsm.config.Debug.FixedRequestID
	} else {
		return "REQ-" + uuid.New().String()
	}
}
