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
}

func NewManager(tafContext core.TafContext, channels core.TafChannels, tlee tleeinterface.TLEE) (*Manager, error) {
	tam := &Manager{
		config:       tafContext.Configuration,
		tafContext:   tafContext,
		channels:     channels,
		sessions:     make(map[string]session.Session),
		workersToTam: make(chan core.Command, tafContext.Configuration.ChanBufSize),
		logger:       logging.CreateChildLogger(tafContext.Logger, "TAM"),
		crypto:       tafContext.Crypto,
		outbox:       channels.OutgoingMessageChannel,
		tlee:         tlee,
		atlResults:   make(map[string]core.AtlResultSet),
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

func (tam *Manager) createSessionId() string {
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
	sessionId := tam.createSessionId()
	//create Session
	session := session.NewInstance(sessionId, cmd.Sender)
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
	//TODO: remove evidence-related data:
	// - unsubscribe evidence subscriptions bound to this session ID
	// - remove subscription data bound to this session ID

	//TODO: force unsubscription of TAS subscription, if existing

	//TODO: remove TMI(s) associated to this session

	//TODO: remove ATL cache entries for this session

	//TODO: remove session data

	success := "Session with ID '" + cmd.Request.SessionID + "' successfully terminated."
	response := tasmsg.TasTeardownResponse{
		AttestationCertificate: tam.crypto.AttestationCertificate(),
		Error:                  nil,
		Success:                &success,
	}

	bytes, err := communication.BuildResponse(tam.config.Communication.TafEndpoint, messages.TAS_TEARDOWN_RESPONSE, cmd.RequestID, response)
	if err != nil {
		tam.logger.Error("Error marshalling response", "error", err)
	}
	//Send response message
	tam.outbox <- core.NewMessage(bytes, "", cmd.ResponseTopic)
	session.TornDown()
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

	for _, tmiID := range targets {
		atlResultSet, exists := tam.atlResults[tmiID]
		propositions := make([]tasmsg.ResultProposition, 0)

		if exists {
			for propositionID, opinion := range atlResultSet.ATLs() {
				trustDecision := atlResultSet.TrustDecisions()[propositionID]
				atl := make([]tasmsg.FluffyActualTrustworthinessLevel, 0)
				baseRate := opinion.BaseRate()
				belief := opinion.Belief()
				disbelief := opinion.Disbelief()
				uncertainty := opinion.Uncertainty()

				atl = append(atl, tasmsg.FluffyActualTrustworthinessLevel{
					Output: tasmsg.FluffyOutput{
						BaseRate:    &baseRate,
						Belief:      &belief,
						Disbelief:   &disbelief,
						Uncertainty: &uncertainty,
					},
					Type: tasmsg.SubjectiveLogicOpinion,
				})

				projectedProbability := atlResultSet.ProjectedProbabilities()[propositionID]
				atl = append(atl, tasmsg.FluffyActualTrustworthinessLevel{
					Output: tasmsg.FluffyOutput{
						Value: &projectedProbability,
					},
					Type: tasmsg.ProjectedProbability,
				})

				propositions = append(propositions, tasmsg.ResultProposition{
					ActualTrustworthinessLevel: atl,
					PropositionID:              propositionID,
					TrustDecision:              &trustDecision,
				})
			}
		}

		taResponseResults = append(taResponseResults, tasmsg.Result{
			ID:           tmiID,
			Propositions: propositions,
		})

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
		bytes, err := communication.BuildResponse(tam.config.Communication.TafEndpoint, messages.TAS_SUBSCRIBE_RESPONSE, cmd.RequestID, response)
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

	//TODO: implement
}

func (tam *Manager) HandleTasUnsubscribeRequest(cmd command.HandleSubscriptionRequest[tasmsg.TasUnsubscribeRequest]) {

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
