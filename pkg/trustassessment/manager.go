package trustassessment

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/vs-uulm/go-taf/internal/flow/completionhandler"
	logging "github.com/vs-uulm/go-taf/internal/logger"
	"github.com/vs-uulm/go-taf/pkg/command"
	"github.com/vs-uulm/go-taf/pkg/communication"
	"github.com/vs-uulm/go-taf/pkg/config"
	"github.com/vs-uulm/go-taf/pkg/core"
	"github.com/vs-uulm/go-taf/pkg/crypto"
	"github.com/vs-uulm/go-taf/pkg/listener"
	"github.com/vs-uulm/go-taf/pkg/manager"
	messages "github.com/vs-uulm/go-taf/pkg/message"
	aivmsg "github.com/vs-uulm/go-taf/pkg/message/aiv"
	mbdmsg "github.com/vs-uulm/go-taf/pkg/message/mbd"
	taqimsg "github.com/vs-uulm/go-taf/pkg/message/taqi"
	tasmsg "github.com/vs-uulm/go-taf/pkg/message/tas"
	tchmsg "github.com/vs-uulm/go-taf/pkg/message/tch"
	v2xmsg "github.com/vs-uulm/go-taf/pkg/message/v2x"
	"github.com/vs-uulm/go-taf/pkg/trustmodel/session"
	"github.com/vs-uulm/taf-tlee-interface/pkg/tleeinterface"
	"hash/fnv"
	"log/slog"
	"slices"
	"strings"
	"time"
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
	//Queryable table of all TMIs
	tmiTable         *TrustModelInstanceTable
	sessionListeners map[listener.SessionListener]bool
	atlListeners     map[listener.ActualTrustLevelListener]bool
	tmiListeners     map[listener.TrustModelInstanceListener]bool
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
		tmiTable:                    CreateTrustModelInstanceTable(),
		sessionListeners:            make(map[listener.SessionListener]bool),
		atlListeners:                make(map[listener.ActualTrustLevelListener]bool),
		tmiListeners:                make(map[listener.TrustModelInstanceListener]bool),
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
		worker := tam.SpawnNewWorker(i, channel, tam.workersToTam, tam.tafContext, tam.tlee, tam.tmiListeners)
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
			case command.HandleRequest[taqimsg.TaqiQuery]:
				tam.HandleTaqiQuery(cmd)
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
			case command.HandleNotify[tchmsg.TchNotify]:
				tmm.HandleTchNotify(cmd) //handle potential trigger based on trustee
				tsm.HandleTchNotify(cmd) //handle evidence from TCH
			// TMM Message Handling
			case command.HandleOneWay[v2xmsg.V2XCpm]:
				tmm.HandleV2xCpmMessage(cmd)
			case command.HandleRequest[tasmsg.TasTmtDiscover]:
				tmm.HandleTasTmtDiscover(cmd)
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
	tam.logger.Debug("Received TAS_INIT command", "Trust Model Template", cmd.Request.TrustModelTemplate, "Client", cmd.Sender)

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
			return
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
	newSession := session.NewInstance(sessionId, cmd.Sender, tmt)
	//put session into session map
	tam.sessions[sessionId] = newSession

	tam.logger.Info("Session created:", "Session ID", newSession.ID(), "Client", newSession.Client())

	//create new TMI and/or dynamic spawn function for session
	tsqs, tMI, dynamicSpawner, err := tmt.Spawn(cmd.Request.Params, tam.tafContext)
	if err != nil {
		delete(tam.sessions, sessionId)
		sendErrorResponse("Error initializing session: " + err.Error())
		return
	} else {
		newSession.SetTrustSourceQuantifiers(tsqs)
		if tMI != nil {
			//add new TMI to session
			sessionTMIs := newSession.TrustModelInstances()
			sessionTMIs[tMI.ID()] = core.MergeFullTMIIdentifier(newSession.Client(), newSession.ID(), newSession.TrustModelTemplate().Identifier(), tMI.ID())
		}
		if dynamicSpawner != nil {
			//register spawn function at session
			newSession.SetDynamicSpawner(dynamicSpawner)

			//iterate over existing triggers and spawn TMIs for these
			if tmt.Type() == core.VEHICLE_TRIGGERED_TRUST_MODEL {
				for _, nodeIdentifier := range tam.tmm.ListRecentV2XNodes() {
					tmi, err := dynamicSpawner.OnNewVehicle(nodeIdentifier, nil)
					if err != nil {
						tam.logger.Error("Error while spawning trust model instance", "TMT", newSession.TrustModelTemplate(), "Identifier used for dynamic spawning", nodeIdentifier)
					} else {
						tmi.Initialize(map[string]interface{}{
							"SourceId": nodeIdentifier,
						})
						tam.AddNewTrustModelInstance(tmi, sessionId)
					}
				}
			} else if tmt.Type() == core.TRUSTEE_TRIGGERED_TRUST_MODEL {
				for _, trusteeIdentifier := range tam.tmm.ListRecentTrustees() {
					tmi, err := dynamicSpawner.OnNewTrustee(trusteeIdentifier, nil)
					if err != nil {
						tam.logger.Error("Error while spawning trust model instance", "TMT", newSession.TrustModelTemplate(), "Identifier used for dynamic spawning", trusteeIdentifier)
					} else {
						tmi.Initialize(map[string]interface{}{
							"trusteeId": trusteeIdentifier,
						})
						tam.AddNewTrustModelInstance(tmi, sessionId)
					}

				}
			}
		}
	}

	successHandler := func() {

		if tMI != nil {
			//add new TMI to list of all TMIs of the TAM
			tam.logger.Debug("TMI spawned:", "TMI ID", tMI.ID(), "Session ID", newSession.ID(), "Client", newSession.Client())

			//Initialize TMI
			tMI.Initialize(nil)

			//Dispatch new TMI instance to worker
			fullTmiID := core.MergeFullTMIIdentifier(newSession.Client(), newSession.ID(), newSession.TrustModelTemplate().Identifier(), tMI.ID())
			tmiInitCmd := command.CreateHandleTMIInit(fullTmiID, tMI)
			tam.DispatchToWorker(newSession, tMI.ID(), tmiInitCmd)
		}

		success := "Session with trust model template '" + tmt.Identifier() + "' created."

		response := tasmsg.TasInitResponse{
			AttestationCertificate: tam.crypto.AttestationCertificate(),
			Error:                  nil,
			SessionID:              &sessionId,
			Success:                &success,
		}

		bytes, errr := communication.BuildResponse(tam.config.Communication.TafEndpoint, messages.TAS_INIT_RESPONSE, cmd.RequestID, response)
		if errr != nil {
			tam.logger.Error("Error marshalling response", "error", err)
			return
		}
		//Send response message
		tam.outbox <- core.NewMessage(bytes, "", cmd.ResponseTopic)
		tam.sessions[sessionId].Established()
		tam.notifySessionCreated(tam.sessions[sessionId])
	}
	errorHandler := func(err error) {

		sendErrorResponse("Error initializing session: " + err.Error())
		//Cleanup TMI creation
		if tMI != nil {
			tMI.Cleanup()
			//Cleanup
			fullTMIid := core.MergeFullTMIIdentifier(newSession.Client(), newSession.ID(), newSession.TrustModelTemplate().Identifier(), tMI.ID())
			tam.tmiTable.UnregisterTMI(newSession.Client(), newSession.ID(), newSession.TrustModelTemplate().Identifier(), tMI.ID())
			//signal worker to destroy TMI
			tam.DispatchToWorker(newSession, tMI.ID(), command.CreateHandleTMIDestroy(fullTMIid))
			//remove ATL cache entries for this session
			delete(tam.atlResults, fullTMIid)
		}
		delete(tam.sessions, sessionId)
	}

	ch := completionhandler.New(successHandler, errorHandler)

	//Initialize Trust Source Quantifiers and Subscriptions
	tam.tsm.SubscribeTrustSourceQuantifiers(newSession, ch)

	ch.Execute()
}

func (tam *Manager) HandleTasTeardownRequest(cmd command.HandleRequest[tasmsg.TasTeardownRequest]) {
	tam.logger.Debug("Received TAS_TEARDOWN command", "Session ID", cmd.Request.SessionID, "Client", cmd.Sender)
	currentSession, exists := tam.sessions[cmd.Request.SessionID]
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
			return
		}
		//Send error message
		tam.outbox <- core.NewMessage(bytes, "", cmd.ResponseTopic)
		return
	}

	currentSession.TearingDown()

	ch := completionhandler.New(func() {
		//Do nothing in case of successfull unregistering of trust sources
	}, func(err error) {
		tam.logger.Error("Error while unregistering trust source quantifiers", "Error Message", err.Error(), "Session ID", currentSession.ID(), "TMT", currentSession.TrustModelTemplate().TemplateName())
	})
	//Foreach Trust Model Instance in Session, unregister trust source quantifiers
	tam.tsm.UnsubscribeTrustSourceQuantifiers(currentSession, ch)
	ch.Execute()

	success := "Session with ID '" + cmd.Request.SessionID + "' successfully terminated."
	response := tasmsg.TasTeardownResponse{
		AttestationCertificate: tam.crypto.AttestationCertificate(),
		Error:                  nil,
		Success:                &success,
	}

	//TODO: force unsubscription of TAS subscription, if existing

	for tmiID, fullTMIID := range currentSession.TrustModelInstances() {
		//signal worker to destroy TMI
		tam.DispatchToWorker(currentSession, tmiID, command.CreateHandleTMIDestroy(fullTMIID))
		//remove ATL cache entries for this session
		delete(tam.atlResults, fullTMIID)
		//remove TMI(s) associated to this session
		delete(currentSession.TrustModelInstances(), tmiID)
	}

	//remove session data
	currentSession.TornDown()
	tam.notifySessionTorndown(currentSession)
	tam.logger.Info("Removing session", "Session ID", currentSession.ID(), "Client", currentSession.Client())
	delete(tam.sessions, currentSession.ID())

	bytes, err := communication.BuildResponse(tam.config.Communication.TafEndpoint, messages.TAS_TEARDOWN_RESPONSE, cmd.RequestID, response)
	if err != nil {
		tam.logger.Error("Error marshalling response", "error", err)
		return
	}
	//Send response message
	tam.outbox <- core.NewMessage(bytes, "", cmd.ResponseTopic)
	return
}

func (tam *Manager) HandleTasTaRequest(cmd command.HandleRequest[tasmsg.TasTaRequest]) {
	tam.logger.Debug("Received TAS_TA_REQUEST command", "Session ID", cmd.Request.SessionID, "Client", cmd.Sender)
	sessionID := cmd.Request.SessionID

	sendErrorResponse := func(errMsg string) {
		response := tasmsg.TasTaResponse{
			AttestationCertificate: tam.crypto.AttestationCertificate(),
			Error:                  &errMsg,
			SessionID:              sessionID,
		}
		bytes, err := communication.BuildResponse(tam.config.Communication.TafEndpoint, messages.TAS_TA_RESPONSE, cmd.RequestID, response)
		if err != nil {
			tam.logger.Error("Error marshalling response", "error", err)
			return
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

	queriedTargets := cmd.Request.Query.Filter
	targets := make([]string, 0)
	if len(queriedTargets) == 0 {
		//when no specific target is specified, use all TMIs from session
		for _, fullTmiID := range tmiSession.TrustModelInstances() {
			targets = append(targets, fullTmiID)
		}
	} else {
		//Check whether all specified targets exist. If at least one is missing, return with error
		errors := make([]string, 0)
		tmiIDs := tmiSession.TrustModelInstances()
		for _, target := range queriedTargets {
			if !tmiSession.HasTMI(target) {
				errors = append(errors, "Target ID '"+target+"' not found.")
			} else {
				//replace short TMI ID by long TMI ID
				targets = append(targets, tmiIDs[target])
			}
		}
		if len(errors) > 0 {
			sendErrorResponse(strings.Join(errors, "\n"))
			return
		}
	}

	tam.logger.Debug("TAS_TA_Request Query Targets", "List", fmt.Sprintf("%v", targets))

	if cmd.Request.AllowCache == nil || *cmd.Request.AllowCache == true {
		//Directly send response
		taResponseResults := make([]tasmsg.Result, 0)

		//Iterate over TMI IDs in the Target Set
		for _, fullTmiID := range targets {
			atlResultSet, exists := tam.atlResults[fullTmiID]

			if exists {
				propositions := make([]Proposition, 0)
				for propositionID := range atlResultSet.ATLs() {
					propositions = append(propositions, NewPropositionEntry(atlResultSet, propositionID))
				}
				_, _, _, tmiID := core.SplitFullTMIIdentifier(fullTmiID)
				result := ResultEntry{
					TmiID:        tmiID,
					Propositions: propositions,
				}
				taResponseResults = append(taResponseResults, result.toTarResultMsgStruct())
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
			return
		}
		tam.outbox <- core.NewMessage(bytes, "", cmd.ResponseTopic)
	} else {
		if tam.sessions[sessionID].TrustModelTemplate().Type() == core.STATIC_TRUST_MODEL &&
			(slices.Contains(tam.sessions[sessionID].TrustModelTemplate().EvidenceTypes(), core.AIV_APPLICATION_ISOLATION) ||
				slices.Contains(tam.sessions[sessionID].TrustModelTemplate().EvidenceTypes(), core.AIV_ACCESS_CONTROL) ||
				slices.Contains(tam.sessions[sessionID].TrustModelTemplate().EvidenceTypes(), core.AIV_CONFIGURATION_INTEGRITY_VERIFICATION) ||
				slices.Contains(tam.sessions[sessionID].TrustModelTemplate().EvidenceTypes(), core.AIV_CONTROL_FLOW_INTEGRITY) ||
				slices.Contains(tam.sessions[sessionID].TrustModelTemplate().EvidenceTypes(), core.AIV_SECURE_OTA) ||
				slices.Contains(tam.sessions[sessionID].TrustModelTemplate().EvidenceTypes(), core.AIV_SECURE_BOOT)) {
			//We need to call AIV Req first and wait for a response
			if len(tam.sessions[sessionID].TrustModelInstances()) > 0 {
				tam.tsm.DispatchAivRequest(tmiSession)
				//Hacky way to emulate allowCache: dispatch AIV Request, then replay TAS_TA_REQUEST after 80 msec - hoping that the AIV Response has been delivered in the meantime
				go func() {
					time.Sleep(80 * time.Millisecond)
					allowCachedNow := true
					cmd.Request.AllowCache = &allowCachedNow
					tam.channels.TAMChannel <- cmd
				}()
			} else {
				sendErrorResponse("No trust model instances found in this session")
				return
			}
		} else {
			sendErrorResponse("Trust model template type does not allow non-cached requests.")
			return
		}
	}

}

func (tam *Manager) HandleTasSubscribeRequest(cmd command.HandleSubscriptionRequest[tasmsg.TasSubscribeRequest]) {
	tam.logger.Debug("Received TAS_SUBSCRIBE_REQUEST command", "Session ID", cmd.Request.SessionID, "Client", cmd.Sender)
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
			return
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
	/*
			TODO: check whether correct IDs are used (short vs full TMI IDs)
		     * ATL Map uses full IDs
			 * Client will use short IDs
	*/
	queriedFilterTargets := cmd.Request.Subscribe.Filter
	filterTargets := make([]string, 0)
	if len(queriedFilterTargets) > 0 {
		//Check whether all specified targets exist. If at least one is missing, return with error
		errors := make([]string, 0)
		fullTMIs := tmiSession.TrustModelInstances()
		for _, target := range queriedFilterTargets {
			if !tmiSession.HasTMI(target) {
				errors = append(errors, "Target ID '"+target+"' not found.")
			} else {
				filterTargets = append(filterTargets, fullTMIs[target])
			}
		}
		if len(errors) > 0 {
			sendErrorResponse(strings.Join(errors, "\n"))
			return
		}
	}

	subscriptionID := tam.generateSubscriptionID()

	subscription := NewSubscription(subscriptionID, sessionID, cmd.SubscriberTopic, filterTargets, trigger)
	tam.tasSubscriptionsToSessionID[subscriptionID] = sessionID
	tam.tasSubscriptions[subscriptionID] = subscription
	//add to session
	tmiSession.AddSubscription(subscriptionID)

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
		return
	}
	tam.outbox <- core.NewMessage(bytes, "", cmd.ResponseTopic)
	tam.logger.Debug("TAS Subscription started", "Session ID", sessionID, "Subscription ID", subscriptionID)

	//Prepare initial TAS_NOTIFY
	var targets []string
	copy(filterTargets, targets)
	if len(targets) == 0 {
		//when no specific target is specified, use all TMIs from session
		for _, tmiID := range tmiSession.TrustModelInstances() {
			targets = append(targets, tmiID)
		}
	}

	taResponseResults := make([]tasmsg.Update, 0)

	//Iterate over TMI IDs in the Target Set
	for _, fullTMIID := range targets {
		atlResultSet, exists := tam.atlResults[fullTMIID]

		if exists {
			propositions := make([]Proposition, 0)
			for propositionID := range atlResultSet.ATLs() {
				propositions = append(propositions, NewPropositionEntry(atlResultSet, propositionID))
			}
			_, _, _, tmiID := core.SplitFullTMIIdentifier(fullTMIID)
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
		tam.logger.Error("Error marshalling notification", "error", err)
		return
	}
	tam.outbox <- core.NewMessage(bytes, "", cmd.SubscriberTopic)
}

func (tam *Manager) HandleTasUnsubscribeRequest(cmd command.HandleSubscriptionRequest[tasmsg.TasUnsubscribeRequest]) {
	tam.logger.Debug("Received TAS_UNSUBSCRIBE_REQUEST command", "Subscription ID", cmd.Request.SubscriptionID, "Session ID", cmd.Request.SessionID, "Client", cmd.Sender)
	sessionID := cmd.Request.SessionID
	subscriptionID := cmd.Request.SubscriptionID

	sendErrorResponse := func(errMsg string) {
		response := tasmsg.TasUnsubscribeResponse{
			AttestationCertificate: tam.crypto.AttestationCertificate(),
			Error:                  &errMsg,
			SessionID:              sessionID,
			Success:                nil,
		}
		bytes, err := communication.BuildResponse(tam.config.Communication.TafEndpoint, messages.TAS_SUBSCRIBE_RESPONSE, cmd.RequestID, response)
		if err != nil {
			tam.logger.Error("Error marshalling response", "error", err)
			return
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
	//delete from session
	tmiSession.RemoveSubscription(subscriptionID)

	//send TAS_UNSUBSCRIBE_RESPONSE
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
		return
	}
	tam.outbox <- core.NewMessage(bytes, "", cmd.ResponseTopic)
	tam.logger.Debug("TAS Subscription terminated", "Session ID", sessionID, "Subscription ID", subscriptionID)
}

func (tam *Manager) HandleTaqiQuery(cmd command.HandleRequest[taqimsg.TaqiQuery]) {
	tam.logger.Debug("Received TAQI_QUERY command", "Target Template", cmd.Request.Query.Template, "Target Identifier", cmd.Request.Query.Identifier, "Target Propositions", cmd.Request.Query.Propositions)
	targetTemplate := cmd.Request.Query.Template
	targetIdentifier := cmd.Request.Query.Identifier
	targetPropositions := cmd.Request.Query.Propositions

	sendErrorResponse := func(errorMsg string) {
		bytes, err := communication.BuildResponse(tam.config.Communication.TafEndpoint, messages.TAQI_RESULT, cmd.RequestID, taqimsg.TaqiResult{
			Error: &errorMsg,
		})
		if err != nil {
			tam.logger.Error("Error marshalling response", "error", err)
			return
		}
		tam.outbox <- core.NewMessage(bytes, "", cmd.ResponseTopic)
	}

	//Directly abort with error result in case the TMT is unknown.
	if nil == tam.tmm.ResolveTMT(targetTemplate) {
		sendErrorResponse("Unknown Trust Model Template used as target: '" + targetTemplate + "'")
		return
	}

	if targetIdentifier == "" {
		sendErrorResponse("Target Identifier must not be empty.")
		return
	}

	tmiQuery := fmt.Sprintf("//*/*/%s/%s", targetTemplate, targetIdentifier)

	matches, err := tam.QueryTMIs(tmiQuery)
	if err != nil {
		sendErrorResponse("Internal error while processing TAQI query.")
		tam.logger.Warn("Internal error while processing TAQI query.", "Error", err.Error())
		return
	}

	//create a map of all propositions to be included in the result set
	targetPropositionMap := make(map[string]struct{})
	for _, proposition := range targetPropositions {
		targetPropositionMap[proposition] = struct{}{}
	}

	taqiResults := make([]taqimsg.Result, 0)
	for _, fullTmiID := range matches {
		atlResultSet, exists := tam.atlResults[fullTmiID] //get cached ATL entry from TMI using the full ID
		if exists {
			propositions := make([]Proposition, 0)
			for propositionID := range atlResultSet.ATLs() {
				//To include the proposition, the list of targetted props must either be empty, or the proposition at hand must be included.
				if _, propExistAsTarget := targetPropositionMap[propositionID]; propExistAsTarget || len(targetPropositions) == 0 {
					propositions = append(propositions, NewPropositionEntry(atlResultSet, propositionID))
				}
			}
			_, _, _, tmiID := core.SplitFullTMIIdentifier(fullTmiID)
			result := ResultEntry{
				TmiID:        tmiID,
				Propositions: propositions,
			}
			taqiResults = append(taqiResults, result.toTaqiResultMsgStruct())
		}
	}
	response := taqimsg.TaqiResult{
		Results: taqiResults,
	}

	bytes, err := communication.BuildResponse(tam.config.Communication.TafEndpoint, messages.TAQI_RESULT, cmd.RequestID, response)
	if err != nil {
		tam.logger.Error("Error marshalling response", "error", err)
		return
	}
	tam.outbox <- core.NewMessage(bytes, "", cmd.ResponseTopic)
}

func (tam *Manager) HandleATLUpdate(cmd command.HandleATLUpdate) {
	tam.logger.Debug("ATL Update", "ResultSet", fmt.Sprintf("%+v", cmd.ResultSet))
	_, sessionID, _, _ := core.SplitFullTMIIdentifier(cmd.FullTmiID)

	_, exists := tam.sessions[sessionID]
	if !exists {
		tam.logger.Debug("ATL Update for unknown session received", "sessionID", sessionID)
		return
	}

	//Check whether there are subscriptions for which the changes are relevant and send out notifications to subscribers
	for _, subscriptionID := range tam.sessions[sessionID].ListSubscriptions() {
		results := tam.tasSubscriptions[subscriptionID].HandleUpdate(tam.atlResults[cmd.FullTmiID], cmd.ResultSet)
		if len(results) > 0 {
			taResponseResults := make([]tasmsg.Update, 0)

			for _, result := range results {
				taResponseResults = append(taResponseResults, result.toUpdateMsgStruct())
			}

			notify := tasmsg.TasNotify{
				AttestationCertificate: tam.crypto.AttestationCertificate(),
				SessionID:              sessionID,
				SubscriptionID:         subscriptionID,
				Updates:                taResponseResults,
			}

			bytes, err := communication.BuildOneWayMessage(tam.config.Communication.TafEndpoint, messages.TAS_NOTIFY, notify)
			if err != nil {
				tam.logger.Error("Error marshalling notification", "error", err)
				return
			}
			tam.outbox <- core.NewMessage(bytes, "", tam.tasSubscriptions[subscriptionID].SubscriberTopic())
		}
	}

	oldATLResults := tam.atlResults[cmd.FullTmiID]

	//overwrite result cache with new values
	tam.atlResults[cmd.FullTmiID] = cmd.ResultSet
	//TODO: make copies of both results and fill cache with new values *before* doing the subscription checks

	tam.notifyATLUpdated(cmd.FullTmiID, oldATLResults, cmd.ResultSet)
}

func (tam *Manager) DispatchToWorker(session session.Session, tmiID string, cmd core.Command) {
	id := core.MergeFullTMIIdentifier(session.Client(), session.ID(), session.TrustModelTemplate().Identifier(), tmiID)
	tam.DispatchToWorkerByFullTMIID(id, cmd)
}

func (tam *Manager) DispatchToWorkerByFullTMIID(fullTMI string, cmd core.Command) {
	workerId := tam.getShardWorkerById(fullTMI)
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

func (tam *Manager) Sessions() map[string]session.Session {
	return tam.sessions
}

func (tam *Manager) AddNewTrustModelInstance(instance core.TrustModelInstance, sessionID string) {
	tmiID := instance.ID()

	//Add TMI to session
	sessions := tam.Sessions()
	sess, exists := sessions[sessionID]
	if !exists {
		tam.logger.Error("Non-existing session used for adding a new TMI", "Session", sessionID, "TMI", instance.ID())
		return

	} else {
		sessionTMIs := sess.TrustModelInstances()
		sessionTMIs[tmiID] = core.MergeFullTMIIdentifier(sess.Client(), sess.ID(), sess.TrustModelTemplate().Identifier(), tmiID)
		tam.tmiTable.RegisterTMI(sess.Client(), sess.ID(), sess.TrustModelTemplate().Identifier(), tmiID)
	}

	//init TMI
	fullTmiID := core.MergeFullTMIIdentifier(sess.Client(), sess.ID(), sess.TrustModelTemplate().Identifier(), instance.ID())

	tmiInitCmd := command.CreateHandleTMIInit(fullTmiID, instance)
	tam.DispatchToWorker(sess, tmiID, tmiInitCmd)
}

func (tam *Manager) RemoveTrustModelInstance(fullTMIid string, sessionID string) {
	sessions := tam.Sessions()
	sess, exists := sessions[sessionID]
	if !exists {
		tam.logger.Error("Non-existing session used for removing a TMI", "Session", sessionID, "TMI", fullTMIid)
		return
	} else {
		_, _, _, tmiID := core.SplitFullTMIIdentifier(fullTMIid)
		tam.logger.Debug("Removing TMI from Session", "Session", sessionID, "TMI", fullTMIid)
		tam.DispatchToWorker(sess, fullTMIid, command.CreateHandleTMIDestroy(fullTMIid))
		tam.tmiTable.UnregisterTMI(sess.Client(), sess.ID(), sess.TrustModelTemplate().Identifier(), tmiID)
		delete(tam.atlResults, fullTMIid)
		tam.notifyATLRemoved(fullTMIid)
		delete(sess.TrustModelInstances(), tmiID)
	}
}

func (tam *Manager) QueryTMIs(query string) ([]string, error) {
	return tam.tmiTable.QueryTMIs(query)
}

func (tam *Manager) AddSessionListener(listener listener.SessionListener) {
	tam.sessionListeners[listener] = true
}

func (tam *Manager) RemoveSessionListener(listener listener.SessionListener) {
	delete(tam.sessionListeners, listener)
}

func (tam *Manager) notifySessionCreated(session session.Session) {
	if len(tam.sessionListeners) > 0 {
		event := listener.NewSessionCreatedEvent(session.ID(), session.TrustModelTemplate(), session.Client())
		for listener, _ := range tam.sessionListeners {
			listener.OnSessionCreated(event)
		}
	}
}

func (tam *Manager) notifySessionTorndown(session session.Session) {
	if len(tam.sessionListeners) > 0 {
		event := listener.NewSessionTorndownEvent(session.ID(), session.TrustModelTemplate(), session.Client())
		for listener, _ := range tam.sessionListeners {
			listener.OnSessionTorndown(event)
		}
	}
}

func (tam *Manager) AddATLListener(listener listener.ActualTrustLevelListener) {
	tam.atlListeners[listener] = true
}

func (tam *Manager) RemoveATLListener(listener listener.ActualTrustLevelListener) {
	delete(tam.atlListeners, listener)
}

func (tam *Manager) notifyATLUpdated(fullTMI string, oldATLs core.AtlResultSet, newATLs core.AtlResultSet) {
	if len(tam.sessionListeners) > 0 {
		event := listener.NewATLUpdatedEvent(fullTMI, newATLs.Version(), oldATLs, newATLs)
		for listener, _ := range tam.atlListeners {
			listener.OnATLUpdated(event)
		}
	}
}

func (tam *Manager) notifyATLRemoved(fullTMI string) {
	if len(tam.sessionListeners) > 0 {
		event := listener.NewATLRemovedEvent(fullTMI)
		for listener, _ := range tam.atlListeners {
			listener.OnATLRemoved(event)
		}
	}
}

func (tam *Manager) AddTMIListener(listener listener.TrustModelInstanceListener) {
	tam.tmiListeners[listener] = true
}
func (tam *Manager) RemoveTMIListener(listener listener.TrustModelInstanceListener) {
	delete(tam.tmiListeners, listener)
}
