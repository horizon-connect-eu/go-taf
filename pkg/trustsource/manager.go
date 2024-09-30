package trustsource

import (
	"errors"
	"fmt"
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
	"log/slog"
	"math"
	"strings"
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

	//Trust Source: Session ID -> TrustSourceQuantifiers
	trustSourceSubscriptions map[core.TrustSource]map[string][]core.TrustSourceQuantifier
	//Trust Source: Trustee ID: Evidence Type -> value
	latestSubscriptionEvidence map[core.TrustSource]map[string]map[core.EvidenceType]int
	//Trust Source: Subscription ID -> true
	subscriptions map[core.TrustSource]map[string]bool
}

func NewManager(tafContext core.TafContext, channels core.TafChannels) (*Manager, error) {
	tsm := &Manager{
		config:                     tafContext.Configuration,
		tafContext:                 tafContext,
		logger:                     logging.CreateChildLogger(tafContext.Logger, "TSM"),
		crypto:                     tafContext.Crypto,
		outbox:                     channels.OutgoingMessageChannel,
		trustSourceSubscriptions:   make(map[core.TrustSource]map[string][]core.TrustSourceQuantifier),
		latestSubscriptionEvidence: make(map[core.TrustSource]map[string]map[core.EvidenceType]int),
		subscriptions:              make(map[core.TrustSource]map[string]bool),
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
		tsm.trustSourceSubscriptions[trustSourceType] = make(map[string][]core.TrustSourceQuantifier)
		tsm.latestSubscriptionEvidence[trustSourceType] = make(map[string]map[core.EvidenceType]int)
		tsm.subscriptions[trustSourceType] = make(map[string]bool)
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

	//Check whether subscription is known, otherwise return
	_, exists := tsm.subscriptions[core.AIV][cmd.Notify.SubscriptionID]
	if !exists {
		tsm.logger.Warn("Unknown subscription for AIV_NOTIFY, discarding message", "Subscription ID", cmd.Notify.SubscriptionID)
		return
	}

	//Extract raw evidence from the message and store it into latestEvidence
	updatedTrustees := make(map[string]bool)
	for _, trusteeReport := range cmd.Notify.TrusteeReports {
		trusteeID := *trusteeReport.TrusteeID
		//Discard old evidence and always create a new map
		tsm.latestSubscriptionEvidence[core.AIV][trusteeID] = make(map[core.EvidenceType]int)
		for _, attestationReport := range trusteeReport.AttestationReport {
			evidenceType := core.EvidenceTypeByName(attestationReport.Claim)
			value := int(attestationReport.Appraisal)
			tsm.latestSubscriptionEvidence[core.AIV][trusteeID][evidenceType] = value
			updatedTrustees[trusteeID] = true
		}
	}

	//Iterate over all sessions register for AIV, call quantifiers and relay updates
	for sessionId, tsqs := range tsm.trustSourceSubscriptions[core.AIV] {
		sess := tsm.tam.Sessions()[sessionId]
		//loop through all updated trustees and find fitting quantifiers; if successful, apply quantifier and add ATO update
		updates := make([]core.Update, 0)
		for trustee := range updatedTrustees {
			for _, tsq := range tsqs {
				if tsq.Trustee == trustee { //TODO
					ato := tsq.Quantifier(tsm.latestSubscriptionEvidence[core.AIV][trustee])
					tsm.logger.Debug("Opinion for "+trustee, "SL", ato.String(), "Input", fmt.Sprintf("%v", tsm.latestSubscriptionEvidence[core.AIV][trustee]))
					updates = append(updates, trustmodelupdate.CreateAtomicTrustOpinionUpdate(ato, "", trustee, core.AIV))
				}
			}
		}
		if len(updates) > 0 {
			for tmiID, fullTmiID := range sess.TrustModelInstances() {
				tmiUpdateCmd := command.CreateHandleTMIUpdate(fullTmiID, updates...)
				tsm.tam.DispatchToWorker(sess, tmiID, tmiUpdateCmd)
			}
		}
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

	//Check whether subscription is known, otherwise return
	_, exists := tsm.subscriptions[core.MBD][cmd.Notify.SubscriptionID]
	if !exists {
		tsm.logger.Warn("Unknown subscription for MBD_NOTIFY, discarding message", "Subscription ID", cmd.Notify.SubscriptionID)
		return
	}

	updatedTrustees := make(map[string]bool)
	sourceID := cmd.Notify.CpmReport.Content.V2XPduEvidence.SourceID
	for _, observation := range cmd.Notify.CpmReport.Content.ObservationSet {
		id := fmt.Sprintf("C_%d_%d", int(sourceID), int(observation.TargetID))
		//Discard old evidence and always create a new map
		tsm.latestSubscriptionEvidence[core.MBD][id] = make(map[core.EvidenceType]int)
		tsm.latestSubscriptionEvidence[core.MBD][id][core.MBD_MISBEHAVIOR_REPORT] = int(observation.Check)
		updatedTrustees[id] = true
	}

	//Iterate over all sessions register for MBD, call quantifiers and relay updates
	for sessionId, tsqs := range tsm.trustSourceSubscriptions[core.MBD] {
		sess := tsm.tam.Sessions()[sessionId]
		//loop through all updated trustees and find fitting quantifiers; if successful, apply quantifier and add ATO update
		updates := make([]core.Update, 0)
		for trustee := range updatedTrustees {
			for _, tsq := range tsqs {
				if tsq.Trustor == "V_ego" && tsq.Trustee == "C_*_*" {
					ato := tsq.Quantifier(tsm.latestSubscriptionEvidence[core.MBD][trustee])
					tsm.logger.Debug("Opinion for "+trustee, "SL", ato.String(), "Input", fmt.Sprintf("%v", tsm.latestSubscriptionEvidence[core.MBD][trustee]))
					updates = append(updates, trustmodelupdate.CreateAtomicTrustOpinionUpdate(ato, "", trustee, core.MBD))
				}
			}
		}
		if len(updates) > 0 {
			for tmiID, fullTmiID := range sess.TrustModelInstances() {
				tmiUpdateCmd := command.CreateHandleTMIUpdate(fullTmiID, updates...)
				tsm.tam.DispatchToWorker(sess, tmiID, tmiUpdateCmd)
			}
		}
	}
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

	//Extract raw evidence from the message and store it into latestEvidence
	trusteeID := cmd.Notify.TchReport.TrusteeID
	updatedTrustees := make(map[string]bool)
	for _, trusteeReport := range cmd.Notify.TchReport.TrusteeReports {
		componentID := trusteeReport.ComponentID
		id := trusteeID
		if componentID != nil {
			id = fmt.Sprintf("%s_%s", trusteeID, *componentID)
		}
		//Discard old evidence and always create a new map
		tsm.latestSubscriptionEvidence[core.TCH][id] = make(map[core.EvidenceType]int)
		for _, attestationReport := range trusteeReport.AttestationReport {
			evidenceType := core.EvidenceTypeByName(attestationReport.Claim)
			value := int(attestationReport.Appraisal)
			tsm.latestSubscriptionEvidence[core.TCH][id][evidenceType] = value
			updatedTrustees[id] = true
		}
	}

	//Iterate over all sessions register for TCH, call quantifiers and relay updates
	for sessionId, tsqs := range tsm.trustSourceSubscriptions[core.TCH] {
		sess := tsm.tam.Sessions()[sessionId]
		//loop through all updated trustees and find fitting quantifiers; if successful, apply quantifier and add ATO update
		updates := make([]core.Update, 0)
		for trustee := range updatedTrustees {
			for _, tsq := range tsqs {
				if tsq.Trustor == "V_ego" && tsq.Trustee == "V_*" {
					ato := tsq.Quantifier(tsm.latestSubscriptionEvidence[core.TCH][trustee])
					tsm.logger.Debug("Opinion for "+trustee, "SL", ato.String(), "Input", fmt.Sprintf("%v", tsm.latestSubscriptionEvidence[core.TCH][trustee]))
					updates = append(updates, trustmodelupdate.CreateAtomicTrustOpinionUpdate(ato, "V_ego", "V_"+trustee, core.TCH))
				}
			}
		}
		if len(updates) > 0 {
			for tmiID, fullTmiID := range sess.TrustModelInstances() {
				tmiUpdateCmd := command.CreateHandleTMIUpdate(fullTmiID, updates...)
				tsm.tam.DispatchToWorker(sess, tmiID, tmiUpdateCmd)
			}
		}
	}
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
			tsm.addSessionToAivSubscription(sess, handler)
		case core.MBD:
			tsm.addSessionToMbdSubscription(sess, handler)
		case core.TCH:
			tsm.addSessionToTchSubscription(sess, handler)
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
			tsm.removeSessionFromAivSubscription(sess, handler)
		case core.MBD:
			tsm.removeSessionFromMbdSubscription(sess, handler)
		case core.TCH:
			tsm.removeSessionFromTchSubscription(sess, handler)
		default:
			//nothing to do
		}
	}
}

func (tsm *Manager) addSessionToTrustSourceSubscription(session session.Session, handler *completionhandler.CompletionHandler, trustSource core.TrustSource) {
	tsqs := make([]core.TrustSourceQuantifier, 0)
	for _, tsq := range session.TrustSourceQuantifiers() {
		if tsq.TrustSource == trustSource {
			tsqs = append(tsqs, tsq)
		}
	}
	tsm.trustSourceSubscriptions[trustSource][session.ID()] = tsqs
}

func (tsm *Manager) addSessionToAivSubscription(session session.Session, handler *completionhandler.CompletionHandler) {
	tsm.addSessionToTrustSourceSubscription(session, handler, core.AIV)
	//TODO: make more robust in case of concurrent subscribe operations
	//TODO: make session-specific subscriptions
	if len(tsm.trustSourceSubscriptions[core.AIV]) == 1 {
		tsm.subscribeAIV(handler)
	}
}
func (tsm *Manager) addSessionToMbdSubscription(session session.Session, handler *completionhandler.CompletionHandler) {
	tsm.addSessionToTrustSourceSubscription(session, handler, core.MBD)
	//TODO: make more robust in case of concurrent subscribe operations
	if len(tsm.trustSourceSubscriptions[core.MBD]) == 1 {
		tsm.subscribeMBD(handler)
	}
}
func (tsm *Manager) addSessionToTchSubscription(session session.Session, handler *completionhandler.CompletionHandler) {
	tsm.addSessionToTrustSourceSubscription(session, handler, core.TCH)
}

func (tsm *Manager) removeSessionFromAivSubscription(session session.Session, handler *completionhandler.CompletionHandler) {
	delete(tsm.trustSourceSubscriptions[core.AIV], session.ID())
	//TODO: implement more robustly
	if len(tsm.trustSourceSubscriptions[core.AIV]) == 0 {
		var subId string
		for key := range tsm.subscriptions[core.AIV] {
			subId = key
			break
		}
		tsm.unsubscribeAIV(subId, handler)
	}
}

func (tsm *Manager) removeSessionFromMbdSubscription(session session.Session, handler *completionhandler.CompletionHandler) {
	delete(tsm.trustSourceSubscriptions[core.MBD], session.ID())
	//TODO: implement more robustly
	if len(tsm.trustSourceSubscriptions[core.MBD]) == 0 {
		var subId string
		for key := range tsm.subscriptions[core.MBD] {
			subId = key
			break
		}
		tsm.unsubscribeMBD(subId, handler)
	}
}

func (tsm *Manager) removeSessionFromTchSubscription(session session.Session, handler *completionhandler.CompletionHandler) {
	delete(tsm.trustSourceSubscriptions[core.TCH], session.ID())
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
						tsm.logger.Debug("Received evidence response from AIV", "Evidence Type", core.EvidenceTypeByName(evidence).String(), "Trustee ID", *trusteeReport.TrusteeID)
						evidenceCollection[core.EvidenceTypeByName(evidence)] = int(report.Appraisal)
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

func (tsm *Manager) subscribeAIV(handler *completionhandler.CompletionHandler) {

	trustees := make(map[string][]core.EvidenceType)
	for _, session := range tsm.tam.Sessions() { //TODO: potential bug if some TMTs are not yet in use but require AIV as well
		for _, tsq := range session.TrustSourceQuantifiers() {
			for _, evidence := range tsq.Evidence {
				if trustees[tsq.Trustee] == nil {
					trustees[tsq.Trustee] = make([]core.EvidenceType, 0)
				}
				trustees[tsq.Trustee] = append(trustees[tsq.Trustee], evidence)
			}
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
			tsm.subscriptions[core.AIV][*cmd.Response.SubscriptionID] = true
			resolve()
		default:
			reject(errors.New("Unknown response type: " + cmd.Type().String()))
		}
	})
	//Send response message
	tsm.outbox <- core.NewMessage(bytes, "", tsm.config.Communication.AivEndpoint)
}

func (tsm *Manager) subscribeMBD(handler *completionhandler.CompletionHandler) {
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
				reject(errors.New(*cmd.Response.Error))
				//TODO: cleanup?
				return
			} else {
				tsm.subscriptions[core.MBD][*cmd.Response.SubscriptionID] = true
				resolve()
			}
		default:
			reject(errors.New("Unknown response type: " + cmd.Type().String()))
		}
	})
	//Send response message
	tsm.outbox <- core.NewMessage(bytes, "", tsm.config.Communication.MbdEndpoint)
}

func (tsm *Manager) unsubscribeAIV(subID string, handler *completionhandler.CompletionHandler) {
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
			//delete associated data structures/lookups
			delete(tsm.subscriptions[core.AIV], subID)

			tsm.logger.Debug("Unregistering AIV Subscription " + subID)

			resolve()
		default:
			reject(errors.New("Unknown response type: " + cmd.Type().String()))
		}
	})
	tsm.outbox <- core.NewMessage(bytes, "", tsm.config.Communication.AivEndpoint)
}

func (tsm *Manager) unsubscribeMBD(subID string, handler *completionhandler.CompletionHandler) {
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
			//delete associated data structures/lookups
			delete(tsm.subscriptions[core.MBD], subID)

			tsm.logger.Debug("Unregistering MBD Subscription " + subID)

			resolve()
		default:
			reject(errors.New("Unknown response type: " + cmd.Type().String()))
		}
	})
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
