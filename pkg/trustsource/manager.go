package trustsource

import (
	"errors"
	"github.com/google/uuid"
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
	"github.com/vs-uulm/go-taf/pkg/trustsource/trustsourcehandler"
	"log/slog"
	"strings"
)

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

	aivHandler *trustsourcehandler.AivHandler
	tchHandler *trustsourcehandler.TchHandler
	mbdHandler *trustsourcehandler.MbdHandler
}

func NewManager(tafContext core.TafContext, channels core.TafChannels) (*Manager, error) {
	tsm := &Manager{
		config:     tafContext.Configuration,
		tafContext: tafContext,
		logger:     logging.CreateChildLogger(tafContext.Logger, "TSM"),
		crypto:     tafContext.Crypto,
		outbox:     channels.OutgoingMessageChannel,
	}
	tsm.logger.Info("Initializing Trust Source Manager")

	tsm.pendingMessageCallbacks = map[messages.MessageSchema]map[string]func(cmd core.Command){
		messages.AIV_SUBSCRIBE_RESPONSE:   make(map[string]func(cmd core.Command)),
		messages.AIV_UNSUBSCRIBE_RESPONSE: make(map[string]func(cmd core.Command)),
		messages.MBD_SUBSCRIBE_RESPONSE:   make(map[string]func(cmd core.Command)),
		messages.MBD_UNSUBSCRIBE_RESPONSE: make(map[string]func(cmd core.Command)),
		messages.AIV_RESPONSE:             make(map[string]func(cmd core.Command)),
	}

	return tsm, nil
}

func (tsm *Manager) SetManagers(managers manager.TafManagers) {
	tsm.tam = managers.TAM
	tsm.tmm = managers.TMM

	//With access to the TMM, we can now check the available TMTs for potential trust sources requested later on
	potentialTrustSources := make(map[core.TrustSource]bool)
	for _, tmt := range tsm.tmm.GetAllTMTs() {
		for _, evidence := range tmt.EvidenceTypes() {
			potentialTrustSources[evidence.Source()] = true
		}
	}
	listOfTrustSources := make([]string, 0)
	//For each type of trust source, initialize some data structures
	for trustSourceType := range potentialTrustSources {
		if trustSourceType == core.AIV {
			tsm.aivHandler = trustsourcehandler.CreateAivHandler(tsm.tam, tsm, tsm.logger)
		}

		if trustSourceType == core.MBD {
			tsm.mbdHandler = trustsourcehandler.CreateMbdHandler(tsm.tam, tsm, tsm.logger)
		}

		if trustSourceType == core.TCH {
			tsm.tchHandler = trustsourcehandler.CreateTchHandler(tsm.tam, tsm.logger)
		}

		listOfTrustSources = append(listOfTrustSources, trustSourceType.String())
	}
	tsm.logger.Info("Registering known trust sources", "Trust Sources", strings.Join(listOfTrustSources, ", "))
}

/* ------------ ------------ AIV Message Handling ------------ ------------ */

func (tsm *Manager) HandleAivResponse(cmd command.HandleResponse[aivmsg.AivResponse]) {
	valid, err := tsm.crypto.VerifyAivResponse(&cmd.Response)
	if err != nil {
		tsm.logger.Error("Error verifying AIV_RESPONSE", "Cause", err)
		return
	}
	if !valid {
		if tsm.config.Crypto.IgnoreVerificationResults {
			tsm.logger.Warn("\n********************** WARNING ********************\n*   Ignoring failed AIV_RESPONSE verification     *\n********************** WARNING ********************")
		} else {
			tsm.logger.Warn("AIV_RESPONSE could not be verified, discarding message")
			return
		}
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
		if tsm.config.Crypto.IgnoreVerificationResults {
			tsm.logger.Warn("\n********************** WARNING **********************\n*   Ignoring failed AIV_NOTIFY verification     *\n************************ WARNING ********************")
		} else {
			tsm.logger.Warn("AIV_NOTIFY could not be verified, discarding message")
			return
		}
	}
	tsm.aivHandler.HandleNotify(cmd)
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
	tsm.mbdHandler.HandleNotify(cmd)
}

/* ------------ ------------ TCH Message Handling ------------ ------------ */

func (tsm *Manager) HandleTchNotify(cmd command.HandleNotify[tchmsg.TchNotify]) {
	valid, err := tsm.crypto.VerifyTchNotify(&cmd.Notify)
	if err != nil {
		tsm.logger.Error("Error verifying TCH_NOTIFY", "Cause", err)
		return
	}
	if !valid {
		if tsm.config.Crypto.IgnoreVerificationResults {
			tsm.logger.Warn("\n********************** WARNING **********************\n*   Ignoring failed TCH_NOTIFY verification     *\n************************ WARNING ********************")
		} else {
			tsm.logger.Warn("TCH_NOTIFY could not be verified, discarding message")
			return
		}
	}
	tsm.tchHandler.HandleNotify(cmd)
}

/* ------------ ------------ TrustSourceQuantifier  Handling ------------ ------------ */

func (tsm *Manager) SubscribeTrustSourceQuantifiers(sess session.Session, handler *completionhandler.CompletionHandler) {

	//When no handler has been set, create empty one
	if handler == nil {
		handler = completionhandler.New(func() {}, func(err error) {
		})
		defer handler.Execute()
	}

	trustSources := make(map[core.TrustSource]bool)
	for _, tsq := range sess.TrustSourceQuantifiers() {
		trustSources[tsq.TrustSource] = true
	}

	for trustSource := range trustSources {
		switch trustSource {
		case core.AIV:
			tsm.aivHandler.AddSession(sess, handler)
		case core.MBD:
			tsm.mbdHandler.AddSession(sess, handler)
		case core.TCH:
			tsm.tchHandler.AddSession(sess, handler)
		default:
			//nothing to do
		}
	}
}

func (tsm *Manager) UnsubscribeTrustSourceQuantifiers(sess session.Session, handler *completionhandler.CompletionHandler) {
	//When no handler has been set, create empty one
	if handler == nil {
		handler = completionhandler.New(func() {}, func(err error) {
		})
		defer handler.Execute()
	}

	trustSources := make(map[core.TrustSource]bool)
	for _, tsq := range sess.TrustSourceQuantifiers() {
		trustSources[tsq.TrustSource] = true
	}

	for trustSource := range trustSources {
		switch trustSource {
		case core.AIV:
			tsm.aivHandler.RemoveSession(sess, handler)
		case core.MBD:
			tsm.mbdHandler.RemoveSession(sess, handler)
		case core.TCH:
			tsm.tchHandler.RemoveSession(sess, handler)

		default:
			//nothing to do
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

	query := make(map[core.TrustSource]map[string][]core.EvidenceType)
	quantifiers := make(map[core.TrustSource]core.Quantifier)

	for _, tsq := range session.TrustSourceQuantifiers() {

		quantifiers[tsq.TrustSource] = tsq.Quantifier
		for _, evidence := range tsq.Evidence {
			if query[evidence.Source()] == nil {
				query[evidence.Source()] = make(map[string][]core.EvidenceType)
			}
			if query[evidence.Source()][tsq.Trustee] == nil {
				query[evidence.Source()][tsq.Trustee] = make([]core.EvidenceType, 0)
			}
			query[evidence.Source()][tsq.Trustee] = append(query[evidence.Source()][tsq.Trustee], evidence)

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
			return
		}

		tsm.RegisterCallback(messages.AIV_RESPONSE, reqId, func(recvCmd core.Command) {
			switch cmd := recvCmd.(type) {
			case command.HandleResponse[aivmsg.AivResponse]:

				updates := make([]core.Update, 0)

				for _, trusteeReport := range cmd.Response.TrusteeReports {
					evidenceCollection := make(map[core.EvidenceType]int)
					for _, report := range trusteeReport.AttestationReport {
						evidence := report.Claim
						tsm.logger.Debug("Received evidence response from AIV", "Evidence Type", core.EvidenceTypeBySourceAndName(core.AIV, evidence).String(), "Trustee ID", *trusteeReport.TrusteeID)
						evidenceCollection[core.EvidenceTypeBySourceAndName(core.AIV, evidence)] = int(report.Appraisal)
					}
					//call quantifier
					ato := quantifiers[core.AIV](evidenceCollection)
					tsm.logger.Debug("Opinion for " + *trusteeReport.TrusteeID + ": " + ato.String())
					updates = append(updates, trustmodelupdate.CreateAtomicTrustOpinionUpdate(ato, "", *trusteeReport.TrusteeID, core.AIV))
				}
				//create update operation for all TMIs of session
				for tmiID, fullTmiID := range session.TrustModelInstances() {
					tmiUpdateCmd := command.CreateHandleTMIUpdate(fullTmiID, updates...)
					tsm.tam.DispatchToWorker(session, tmiID, tmiUpdateCmd)
				}
			default:
				//Nothing to do
			}
		})
		//Send response message
		tsm.outbox <- core.NewMessage(bytes, "", tsm.config.Communication.AivEndpoint)
	}
}

func (tsm *Manager) SubscribeAIV(handler *completionhandler.CompletionHandler, sess session.Session) {

	trustees := make(map[string][]core.EvidenceType)
	for _, tsq := range sess.TrustSourceQuantifiers() {
		for _, evidence := range tsq.Evidence {
			if trustees[tsq.Trustee] == nil {
				trustees[tsq.Trustee] = make([]core.EvidenceType, 0)
			}
			trustees[tsq.Trustee] = append(trustees[tsq.Trustee], evidence)
		}
	}

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
		return
	}

	resolve, reject := handler.Register()

	tsm.RegisterCallback(messages.AIV_SUBSCRIBE_RESPONSE, subReqId, func(recvCmd core.Command) {
		switch cmd := recvCmd.(type) {
		case command.HandleResponse[aivmsg.AivSubscribeResponse]:
			if cmd.Response.Error != nil {
				reject(errors.New(*cmd.Response.Error))
				return
			}
			tsm.aivHandler.RegisterSubscription(sess, *cmd.Response.SubscriptionID)
			resolve()
		default:
			reject(errors.New("Unknown response type: " + cmd.Type().String()))
		}
	})
	//Send response message
	tsm.outbox <- core.NewMessage(bytes, "", tsm.config.Communication.AivEndpoint)
}

func (tsm *Manager) SubscribeMBD(handler *completionhandler.CompletionHandler) {
	subMsg := mbdmsg.MBDSubscribeRequest{
		AttestationCertificate: tsm.crypto.AttestationCertificate(),
		Subscribe:              true,
	}
	subReqId := tsm.GenerateRequestId()
	bytes, err := communication.BuildSubscriptionRequest(tsm.config.Communication.TafEndpoint, messages.MBD_SUBSCRIBE_REQUEST, tsm.config.Communication.TafEndpoint, tsm.config.Communication.TafEndpoint, subReqId, subMsg)
	if err != nil {
		tsm.logger.Error("Error marshalling response", "error", err)
		return
	}

	resolve, reject := handler.Register()
	tsm.RegisterCallback(messages.MBD_SUBSCRIBE_RESPONSE, subReqId, func(recvCmd core.Command) {
		switch cmd := recvCmd.(type) {
		case command.HandleResponse[mbdmsg.MBDSubscribeResponse]:
			if cmd.Response.Error != nil {
				tsm.mbdHandler.SetSubscriptionState(trustsourcehandler.NA)
				reject(errors.New(*cmd.Response.Error))
				return
			} else {
				tsm.mbdHandler.SetSubscriptionId(*cmd.Response.SubscriptionID)
				tsm.mbdHandler.SetSubscriptionState(trustsourcehandler.SUBSCRIBED)
				resolve()
			}
		default:
			reject(errors.New("Unknown response type: " + cmd.Type().String()))
		}
	})

	//Send subscription request
	tsm.mbdHandler.SetSubscriptionState(trustsourcehandler.SUBSCRIBING)
	tsm.outbox <- core.NewMessage(bytes, "", tsm.config.Communication.MbdEndpoint)
}

func (tsm *Manager) UnsubscribeAIV(subID string, handler *completionhandler.CompletionHandler) {
	resolve, reject := handler.Register()

	unsubMsg := aivmsg.AivUnsubscribeRequest{
		AttestationCertificate: tsm.crypto.AttestationCertificate(),
		SubscriptionID:         subID,
	}
	unsubReqId := tsm.GenerateRequestId()
	bytes, err := communication.BuildSubscriptionRequest(tsm.config.Communication.TafEndpoint, messages.AIV_UNSUBSCRIBE_REQUEST, tsm.config.Communication.TafEndpoint, tsm.config.Communication.TafEndpoint, unsubReqId, unsubMsg)
	if err != nil {
		tsm.logger.Error("Error marshalling response", "error", err)
		return
	}
	tsm.RegisterCallback(messages.AIV_UNSUBSCRIBE_RESPONSE, unsubReqId, func(recvCmd core.Command) {
		switch cmd := recvCmd.(type) {
		case command.HandleResponse[aivmsg.AivUnsubscribeResponse]:
			if cmd.Response.Error != nil {
				reject(errors.New(*cmd.Response.Error))
				return
			}
			resolve()
		default:
			reject(errors.New("Unknown response type: " + cmd.Type().String()))
		}
	})
	tsm.outbox <- core.NewMessage(bytes, "", tsm.config.Communication.AivEndpoint)
}

func (tsm *Manager) UnsubscribeMBD(subID string, handler *completionhandler.CompletionHandler) {
	resolve, reject := handler.Register()

	unsubMsg := mbdmsg.MBDUnsubscribeRequest{
		AttestationCertificate: tsm.crypto.AttestationCertificate(),
		SubscriptionID:         subID,
	}
	unsubReqId := tsm.GenerateRequestId()
	bytes, err := communication.BuildSubscriptionRequest(tsm.config.Communication.TafEndpoint, messages.MBD_UNSUBSCRIBE_REQUEST, tsm.config.Communication.TafEndpoint, tsm.config.Communication.TafEndpoint, unsubReqId, unsubMsg)
	if err != nil {
		tsm.logger.Error("Error marshalling response", "error", err)
		return
	}
	tsm.RegisterCallback(messages.MBD_UNSUBSCRIBE_RESPONSE, unsubReqId, func(recvCmd core.Command) {
		switch cmd := recvCmd.(type) {
		case command.HandleResponse[mbdmsg.MBDUnsubscribeResponse]:
			if cmd.Response.Error != nil {
				reject(errors.New(*cmd.Response.Error))
				return
			}
			tsm.mbdHandler.SetSubscriptionState(trustsourcehandler.NA)
			tsm.logger.Debug("Unregistering MBD Subscription " + subID)

			resolve()
		default:
			reject(errors.New("Unknown response type: " + cmd.Type().String()))
		}
	})
	tsm.mbdHandler.SetSubscriptionState(trustsourcehandler.UNSUBSCRIBING)
	tsm.outbox <- core.NewMessage(bytes, "", tsm.config.Communication.MbdEndpoint)
}

func (tsm *Manager) GenerateRequestId() string {
	//When debug configuration provides fixed session ID, use that ID
	if tsm.config.Debug.FixedRequestID != "" {
		return tsm.config.Debug.FixedRequestID
	} else {
		return "REQ-" + uuid.New().String()
	}
}
