package trustassessment

import (
	"context"
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
	tasmsg "github.com/vs-uulm/go-taf/pkg/message/tas"
	tchmsg "github.com/vs-uulm/go-taf/pkg/message/tch"
	v2xmsg "github.com/vs-uulm/go-taf/pkg/message/v2x"
	"github.com/vs-uulm/go-taf/pkg/trustmodel/session"
	"github.com/vs-uulm/taf-tlee-interface/pkg/tleeinterface"
	"hash/fnv"
	"log/slog"
	"strings"
)

type Manager struct {
	config       config.Configuration
	tamToWorkers []chan core.Command
	workersToTam chan core.Command
	logger       *slog.Logger
	tafContext   core.TafContext
	channels     core.TafChannels
	//sessionID->Session
	sessions map[string]session.Session
	outbox   chan core.Message
	tsm      manager.TrustSourceManager
	tmm      manager.TrustModelManager
	crypto   *crypto.Crypto
	tlee     tleeinterface.TLEE
	//tmiID->latest ATLs/PPs/TDs
	atlResults map[string]core.AtlResultSet
	//tas sub ID->sessionID
	tasSubscriptionsToSessionID map[string]string
	//tas sub ID->Subscription
	tasSubscriptions map[string]Subscription
}

func NewManager(tafContext core.TafContext, channels core.TafChannels, tlee tleeinterface.TLEE) (*Manager, error) {
	tam := &Manager{
		config:                      tafContext.Configuration,
		tafContext:                  tafContext,
		channels:                    channels,
		sessions:                    make(map[string]session.Session),
		workersToTam:                make(chan core.Command, tafContext.Configuration.ChanBufSize),
		logger:                      logging.CreateChildLogger(tafContext.Logger, "TAM"),
		crypto:                      tafContext.Crypto,
		outbox:                      channels.OutgoingMessageChannel,
		tlee:                        tlee,
		atlResults:                  make(map[string]core.AtlResultSet),
		tasSubscriptionsToSessionID: make(map[string]string),
		tasSubscriptions:            make(map[string]Subscription),
	}
	tam.logger.Info("Initializing Trust Assessment Manager", "Worker Count", tam.config.TAM.TrustModelInstanceShards)
	return tam, nil
}

func (tam *Manager) SetManagers(managers manager.TafManagers) {
	tam.tmm = managers.TMM
	tam.tsm = managers.TSM
}

// Run the trust assessment manager
func (tam *Manager) Run() {

	defer func() {
		tam.logger.Info("Shutting down")
	}()

	tsm := tam.tsm
	tmm := tam.tmm

	tam.tamToWorkers = make([]chan core.Command, 0, tam.config.TAM.TrustModelInstanceShards)
	for i := range tam.config.TAM.TrustModelInstanceShards {
		channel := make(chan core.Command, tam.config.ChanBufSize)
		tam.tamToWorkers = append(tam.tamToWorkers, channel)
		worker := tam.SpawnNewWorker(i, channel, tam.workersToTam, tam.tafContext, tam.tlee)
		go worker.Run()
	}

	for {
		// Each iteration, check whether we've been cancelled.
		if err := context.Cause(tam.tafContext.Context); err != nil {
			return
		}
		select {
		case <-tam.tafContext.Context.Done():
			if len(tam.channels.TAMChannel) != 0 {
				continue
			}
			return
		case incomingCmd := <-tam.workersToTam:
			switch cmd := incomingCmd.(type) {
			case command.HandleATLUpdate:
				tam.HandleATLUpdate(cmd)
			default:
				tam.logger.Warn("Command with no associated handling logic received by TAM from Worker", "Command Type", cmd.Type())
			}
		case incomingCmd := <-tam.channels.TAMChannel:
			switch cmd := incomingCmd.(type) {
			// TAM Message Handling
			case command.HandleRequest[tasmsg.TasInitRequest]:
				tam.HandleTasInitRequest(cmd)
			case command.HandleRequest[tasmsg.TasTeardownRequest]:
				tam.HandleTasTeardownRequest(cmd)
			case command.HandleRequest[tasmsg.TasTaRequest]:
				tam.HandleTasTaRequest(cmd)
			case command.HandleSubscriptionRequest[tasmsg.TasSubscribeRequest]:
				tam.HandleTasSubscribeRequest(cmd)
			case command.HandleSubscriptionRequest[tasmsg.TasUnsubscribeRequest]:
				tam.HandleTasUnsubscribeRequest(cmd)
			// TSM Message Handling
			case command.HandleResponse[aivmsg.AivResponse]:
				tsm.HandleAivResponse(cmd)
			case command.HandleResponse[aivmsg.AivSubscribeResponse]:
				tsm.HandleAivSubscribeResponse(cmd)
			case command.HandleResponse[aivmsg.AivUnsubscribeResponse]:
				tsm.HandleAivUnsubscribeResponse(cmd)
			case command.HandleNotify[aivmsg.AivNotify]:
				tsm.HandleAivNotify(cmd)
			case command.HandleResponse[mbdmsg.MBDSubscribeResponse]:
				tsm.HandleMbdSubscribeResponse(cmd)
			case command.HandleResponse[mbdmsg.MBDUnsubscribeResponse]:
				tsm.HandleMbdUnsubscribeResponse(cmd)
			case command.HandleNotify[mbdmsg.MBDNotify]:
				tsm.HandleMbdNotify(cmd)
			case command.HandleNotify[tchmsg.Message]:
				tsm.HandleTchNotify(cmd)
			// TMM Message Handling
			case command.HandleOneWay[v2xmsg.V2XCpm]:
				tmm.HandleV2xCpmMessage(cmd)
			default:
				tam.logger.Warn("Command with no associated handling logic received by TAM from Communication Handler", "Command Type", cmd.Type())
			}
		}
	}
}

func (tam *Manager) generateSessionId() string {
	//When debug configuration provides fixed session ID, use this ID
	if tam.config.Debug.FixedSessionID != "" {
		return tam.config.Debug.FixedSessionID
	} else {
		return "SES-" + uuid.New().String()
	}
}

func (tam *Manager) HandleTasInitRequest(cmd command.HandleRequest[tasmsg.TasInitRequest]) {
	tam.logger.Info("Received TAS_INIT command", "Trust Model", cmd.Request.TrustModelTemplate)

	sendErrorResponse := func(errorMsg string) {
		response := tasmsg.TasInitResponse{
			AttestationCertificate: tam.crypto.AttestationCertificate(),
			Error:                  &errorMsg,
			SessionID:              nil,
			Success:                nil,
		}
		bytes, err := communication.BuildResponse(tam.config.Communication.TafEndpoint, messages.TAS_INIT_RESPONSE, cmd.RequestID, response)
		if err != nil {
			tam.logger.Error("Error marshalling response", "error", err)
		}
		//Send error message
		tam.outbox <- core.NewMessage(bytes, "", cmd.ResponseTopic)
	}

	tmt := tam.tmm.ResolveTMT(cmd.Request.TrustModelTemplate)
	if tmt == nil {
		tam.logger.Warn("Unknown Trust Model Template or Version:" + cmd.Request.TrustModelTemplate)
		errorMsg := "Trust model template '" + cmd.Request.TrustModelTemplate + "' could not be resolved."
		sendErrorResponse(errorMsg)
		return
	}
	//create session ID for client
	sessionId := tam.generateSessionId()
	//create Session
	session := session.NewInstance(sessionId, cmd.Sender, tmt)
	//put session into session map
	tam.sessions[sessionId] = session

	tam.logger.Info("Session created:", "Session ID", session.ID(), "Client", session.Client())

	//create new TMI for session //TODO: always possible for dynamic models?

	tMI, err := tmt.Spawn(cmd.Request.Params, tam.tafContext, tam.channels)
	if err != nil {
		delete(tam.sessions, sessionId)
		sendErrorResponse("Error initializing session: " + err.Error())
		return
	}
	//add new TMI to session
	sessionTMIs := session.TrustModelInstances()
	sessionTMIs[tMI.ID()] = true
	tmiID := tMI.ID()

	successHandler := func() {
		//add new TMI to list of all TMIs of the TAM
		tam.logger.Info("TMI spawned:", "TMI ID", tMI.ID(), "Session ID", session.ID(), "Client", session.Client())

		//Initialize TMI
		tMI.Initialize(nil)

		//Dispatch new TMI instance to worker
		tmiInitCmd := command.CreateHandleTMIInit(tmiID, tMI, sessionId)
		tam.DispatchToWorker(tmiID, tmiInitCmd)

		success := "Session with trust model template '" + tMI.Template().TemplateName() + "@" + tMI.Template().Version() + "' created."

		response := tasmsg.TasInitResponse{
			AttestationCertificate: tam.crypto.AttestationCertificate(),
			Error:                  nil,
			SessionID:              &sessionId,
			Success:                &success,
		}

		bytes, err := communication.BuildResponse(tam.config.Communication.TafEndpoint, messages.TAS_INIT_RESPONSE, cmd.RequestID, response)
		if err != nil {
			tam.logger.Error("Error marshalling response", "error", err)
		}
		//Send response message
		tam.outbox <- core.NewMessage(bytes, "", cmd.ResponseTopic)
		tam.sessions[sessionId].Established()
	}
	errorHandler := func(err error) {
		//TODO: undo session, TMI, etc.
		sendErrorResponse("Error initializing session: " + err.Error())
		//Cleanup TMI creation
		tMI.Cleanup()
		delete(tam.sessions, sessionId)
	}

	ch := completionhandler.New(successHandler, errorHandler)

	//Initialize Trust Source Quantifiers and Subscriptions
	tam.tsm.RegisterTrustSourceQuantifiers(tmt, tmiID, ch)

	ch.Execute()
}

func (tam *Manager) HandleTasTeardownRequest(cmd command.HandleRequest[tasmsg.TasTeardownRequest]) {
	tam.logger.Info("Received TAS_TEARDOWN command", "Session ID", cmd.Request.SessionID)
	session, exists := tam.sessions[cmd.Request.SessionID]
	if !exists {
		errorMsg := "Session ID '" + cmd.Request.SessionID + "' not found."

		response := tasmsg.TasTeardownResponse{
			AttestationCertificate: tam.crypto.AttestationCertificate(),
			Error:                  &errorMsg,
			Success:                nil,
		}
		bytes, err := communication.BuildResponse(tam.config.Communication.TafEndpoint, messages.TAS_TEARDOWN_RESPONSE, cmd.RequestID, response)
		if err != nil {
			tam.logger.Error("Error marshalling response", "error", err)
		}
		//Send error message
		tam.outbox <- core.NewMessage(bytes, "", cmd.ResponseTopic)
		return
	}

	session.TearingDown()

	ch := completionhandler.New(func() {
		//Do nothing in case of successfull unregistering of trust sources
	}, func(err error) {
		tam.logger.Error("Error while unregistering trust source quantifiers", "Error Message", err.Error(), "Session ID", session.ID(), "TMT", session.TrustModelTemplate().TemplateName())
	})
	//Foreach Trust Model Instance in Session, unregister trust source quantifiers
	for tmiID, _ := range session.TrustModelInstances() {
		tam.tsm.UnregisterTrustSourceQuantifiers(session.TrustModelTemplate(), tmiID, ch)
	}
	ch.Execute()

	success := "Session with ID '" + cmd.Request.SessionID + "' successfully terminated."
	response := tasmsg.TasTeardownResponse{
		AttestationCertificate: tam.crypto.AttestationCertificate(),
		Error:                  nil,
		Success:                &success,
	}

	//TODO: force unsubscription of TAS subscription, if existing

	//signal worker to destroy TMI
	for tmiID, _ := range session.TrustModelInstances() {
		tam.DispatchToWorker(tmiID, command.CreateHandleTMIDestroy(tmiID))
	}

	//remove ATL cache entries for this session
	for tmiID, _ := range session.TrustModelInstances() {
		delete(tam.atlResults, tmiID)
	}

	//remove TMI(s) associated to this session
	for tmiID, _ := range session.TrustModelInstances() {
		delete(session.TrustModelInstances(), tmiID)
	}

	//remove session data
	session.TornDown()
	delete(tam.sessions, session.ID())

	bytes, err := communication.BuildResponse(tam.config.Communication.TafEndpoint, messages.TAS_TEARDOWN_RESPONSE, cmd.RequestID, response)
	if err != nil {
		tam.logger.Error("Error marshalling response", "error", err)
	}
	//Send response message
	tam.outbox <- core.NewMessage(bytes, "", cmd.ResponseTopic)
	return
}

func (tam *Manager) HandleTasTaRequest(cmd command.HandleRequest[tasmsg.TasTaRequest]) {
	sessionID := cmd.Request.SessionID
	allowCached := cmd.Request.AllowCache
	util.UNUSED(allowCached)

	sendErrorResponse := func(errMsg string) {
		response := tasmsg.TasTaResponse{
			AttestationCertificate: tam.crypto.AttestationCertificate(),
			Error:                  &errMsg,
			SessionID:              sessionID,
		}
		bytes, err := communication.BuildResponse(tam.config.Communication.TafEndpoint, messages.TAS_TA_RESPONSE, cmd.RequestID, response)
		if err != nil {
			tam.logger.Error("Error marshalling response", "error", err)
		}
		tam.outbox <- core.NewMessage(bytes, "", cmd.ResponseTopic)
	}

	tmiSession, exists := tam.sessions[sessionID]
	if !exists {
		sendErrorResponse("Unknown session")
		return
	} else if tmiSession.State() != session.ESTABLISHED {
		sendErrorResponse("Session not in established state")
		return
	}

	targets := cmd.Request.Query.Filter
	if len(targets) == 0 {
		//when no specific target is specified, use all TMIs from session
		for tmiID, _ := range tmiSession.TrustModelInstances() {
			targets = append(targets, tmiID)
		}
	} else {
		//Check whether all specified targets exist. If at least one is missing, return with error
		errors := make([]string, 0)
		for _, target := range targets {
			if !tmiSession.HasTMI(target) {
				errors = append(errors, "Target ID '"+target+"' not found.")
			}
		}
		if len(errors) > 0 {
			sendErrorResponse(strings.Join(errors, "\n"))
			return
		}
	}

	taResponseResults := make([]tasmsg.Result, 0)

	//Iterate over TMI IDs in the Target Set
	for _, tmiID := range targets {
		atlResultSet, exists := tam.atlResults[tmiID]

		if exists {
			propositions := make([]Proposition, 0)
			for propositionID, _ := range atlResultSet.ATLs() {
				propositions = append(propositions, NewPropositionEntry(atlResultSet, propositionID))
			}
			result := ResultEntry{
				TmiID:        tmiID,
				Propositions: propositions,
			}
			taResponseResults = append(taResponseResults, result.toResultMsgStruct())
		}
	}

	response := tasmsg.TasTaResponse{
		AttestationCertificate: tam.crypto.AttestationCertificate(),
		Error:                  nil,
		Results:                taResponseResults,
		SessionID:              sessionID,
	}
	bytes, err := communication.BuildResponse(tam.config.Communication.TafEndpoint, messages.TAS_TA_RESPONSE, cmd.RequestID, response)
	if err != nil {
		tam.logger.Error("Error marshalling response", "error", err)
	}
	tam.outbox <- core.NewMessage(bytes, "", cmd.ResponseTopic)
}

func (tam *Manager) HandleTasSubscribeRequest(cmd command.HandleSubscriptionRequest[tasmsg.TasSubscribeRequest]) {
	sessionID := cmd.Request.SessionID

	sendErrorResponse := func(errMsg string) {
		response := tasmsg.TasSubscribeResponse{
			AttestationCertificate: tam.crypto.AttestationCertificate(),
			Error:                  &errMsg,
			SessionID:              sessionID,
			SubscriptionID:         nil,
			Success:                nil,
		}
		bytes, err := communication.BuildSubscriptionResponse(tam.config.Communication.TafEndpoint, messages.TAS_SUBSCRIBE_RESPONSE, cmd.RequestID, response)
		if err != nil {
			tam.logger.Error("Error marshalling response", "error", err)
		}
		tam.outbox <- core.NewMessage(bytes, "", cmd.ResponseTopic)
	}

	tmiSession, exists := tam.sessions[sessionID]
	if !exists {
		sendErrorResponse("Unknown session")
		return
	} else if tmiSession.State() != session.ESTABLISHED {
		sendErrorResponse("Session not in established state")
		return
	}

	//Set trigger type
	var trigger Trigger
	if string(cmd.Request.Trigger) == string(ACTUAL_TRUSTWORTHINESS_LEVEL) {
		trigger = ACTUAL_TRUSTWORTHINESS_LEVEL
	} else if string(cmd.Request.Trigger) == string(TRUST_DECISION) {
		trigger = TRUST_DECISION
	} else {
		sendErrorResponse("Unknown trigger used: " + string(cmd.Request.Trigger))
		return
	}

	//Set filter targets
	filter := cmd.Request.Subscribe.Filter
	if len(filter) > 0 {
		//Check whether all specified targets exist. If at least one is missing, return with error
		errors := make([]string, 0)
		for _, target := range filter {
			if !tmiSession.HasTMI(target) {
				errors = append(errors, "Target ID '"+target+"' not found.")
			}
		}
		if len(errors) > 0 {
			sendErrorResponse(strings.Join(errors, "\n"))
			return
		}
	}

	subscriptionID := tam.generateSubscriptionID()

	subscription := NewSubscription(subscriptionID, sessionID, cmd.SubscriberTopic, filter, trigger)
	tam.tasSubscriptionsToSessionID[subscriptionID] = sessionID
	tam.tasSubscriptions[subscriptionID] = subscription

	//send TAS_SUBSCRIBE_RESPONSE
	success := "Subscription successfully created."
	response := tasmsg.TasSubscribeResponse{
		AttestationCertificate: tam.crypto.AttestationCertificate(),
		Error:                  nil,
		SessionID:              sessionID,
		SubscriptionID:         &subscriptionID,
		Success:                &success,
	}
	bytes, err := communication.BuildSubscriptionResponse(tam.config.Communication.TafEndpoint, messages.TAS_SUBSCRIBE_RESPONSE, cmd.RequestID, response)
	if err != nil {
		tam.logger.Error("Error marshalling response", "error", err)
	}
	tam.outbox <- core.NewMessage(bytes, "", cmd.ResponseTopic)
	tam.logger.Info("Subscription started", "Session ID", sessionID, "Subscription ID", subscriptionID)

	//Prepare initial TAS_NOTIFY
	var targets []string
	copy(filter, targets)
	if len(targets) == 0 {
		//when no specific target is specified, use all TMIs from session
		for tmiID, _ := range tmiSession.TrustModelInstances() {
			targets = append(targets, tmiID)
		}
	}

	taResponseResults := make([]tasmsg.Update, 0)

	//Iterate over TMI IDs in the Target Set
	for _, tmiID := range targets {
		atlResultSet, exists := tam.atlResults[tmiID]

		if exists {
			propositions := make([]Proposition, 0)
			for propositionID, _ := range atlResultSet.ATLs() {
				propositions = append(propositions, NewPropositionEntry(atlResultSet, propositionID))
			}
			result := ResultEntry{
				TmiID:        tmiID,
				Propositions: propositions,
			}
			taResponseResults = append(taResponseResults, result.toUpdateMsgStruct())
		}
	}

	initialNotify := tasmsg.TasNotify{
		AttestationCertificate: tam.crypto.AttestationCertificate(),
		SessionID:              sessionID,
		SubscriptionID:         subscriptionID,
		Updates:                taResponseResults,
	}

	bytes, err = communication.BuildOneWayMessage(tam.config.Communication.TafEndpoint, messages.TAS_NOTIFY, initialNotify)
	if err != nil {
		tam.logger.Error("Error marshalling response", "error", err)
	}
	tam.outbox <- core.NewMessage(bytes, "", cmd.SubscriberTopic)
}

func (tam *Manager) HandleTasUnsubscribeRequest(cmd command.HandleSubscriptionRequest[tasmsg.TasUnsubscribeRequest]) {
	sessionID := cmd.Request.SessionID
	subscriptionID := cmd.Request.SubscriptionID

	sendErrorResponse := func(errMsg string) {
		response := tasmsg.TasSubscribeResponse{
			AttestationCertificate: tam.crypto.AttestationCertificate(),
			Error:                  &errMsg,
			SessionID:              sessionID,
			SubscriptionID:         nil,
			Success:                nil,
		}
		bytes, err := communication.BuildResponse(tam.config.Communication.TafEndpoint, messages.TAS_SUBSCRIBE_RESPONSE, cmd.RequestID, response)
		if err != nil {
			tam.logger.Error("Error marshalling response", "error", err)
		}
		tam.outbox <- core.NewMessage(bytes, "", cmd.ResponseTopic)
	}

	//check whether session exists
	tmiSession, exists := tam.sessions[sessionID]
	if !exists {
		sendErrorResponse("Unknown session with ID '" + sessionID + "'")
		return
	} else if tmiSession.State() != session.ESTABLISHED {
		sendErrorResponse("Session not in established state")
		return
	}

	//check whether subscription exists
	_, exists = tam.tasSubscriptions[subscriptionID]
	if !exists {
		sendErrorResponse("Unknown subscription with ID '" + subscriptionID + "'")
		return
	}

	//unregister subscription handler
	delete(tam.tasSubscriptions, subscriptionID)
	//remove from map
	delete(tam.tasSubscriptionsToSessionID, subscriptionID)

	//send TAS_SUBSCRIBE_RESPONSE
	success := "Subscription with ID '" + subscriptionID + "' successfully terminated."
	response := tasmsg.TasUnsubscribeResponse{
		AttestationCertificate: tam.crypto.AttestationCertificate(),
		Error:                  nil,
		SessionID:              sessionID,
		Success:                &success,
	}
	bytes, err := communication.BuildSubscriptionResponse(tam.config.Communication.TafEndpoint, messages.TAS_UNSUBSCRIBE_RESPONSE, cmd.RequestID, response)
	if err != nil {
		tam.logger.Error("Error marshalling response", "error", err)
	}
	tam.outbox <- core.NewMessage(bytes, "", cmd.ResponseTopic)
	tam.logger.Info("Subscription terminated", "Session ID", sessionID, "Subscription ID", subscriptionID)
}

func (tam *Manager) HandleATLUpdate(cmd command.HandleATLUpdate) {
	tam.logger.Debug("ATL Update", "ResultSet", fmt.Sprintf("%+v", cmd.ResultSet))
	tmiID := cmd.ResultSet.TmiID()
	sessionID := cmd.Session
	_, exists := tam.sessions[sessionID]
	if !exists {
		return
	}
	//TODO: Check whether there are relevant(?) changes and notify potential subscribers
	tam.atlResults[tmiID] = cmd.ResultSet
}

func (tam *Manager) DispatchToWorker(tmiID string, cmd core.Command) {
	workerId := tam.getShardWorkerById(tmiID)
	tam.tamToWorkers[workerId] <- cmd
}

// Get shard worker based on provided ID and configured number of shards
func (tam *Manager) getShardWorkerById(stringID string) int {
	algorithm := fnv.New32a()
	_, err := algorithm.Write([]byte(stringID))
	if err != nil {
		return 0
	} else {
		id := int(algorithm.Sum32())
		return id % tam.config.TAM.TrustModelInstanceShards
	}
}

func (tam *Manager) generateSubscriptionID() string {
	//When debug configuration provides fixed subscription ID, use this ID
	if tam.config.Debug.FixedSubscriptionID != "" {
		return tam.config.Debug.FixedSubscriptionID
	} else {
		return "SUB-" + uuid.New().String()
	}
}
