package trustassessment

import (
	"context"
	"crypto-library-interface/pkg/crypto"
	logging "github.com/vs-uulm/go-taf/internal/logger"
	"github.com/vs-uulm/go-taf/pkg/command"
	"github.com/vs-uulm/go-taf/pkg/communication"
	"github.com/vs-uulm/go-taf/pkg/config"
	"github.com/vs-uulm/go-taf/pkg/core"
	tasmsg "github.com/vs-uulm/go-taf/pkg/message/tas"
	"github.com/vs-uulm/go-taf/pkg/trustmodel"
	"github.com/vs-uulm/go-taf/pkg/trustmodel/session"
	"github.com/vs-uulm/go-taf/pkg/trustmodel/trustmodelinstance"
	"github.com/vs-uulm/go-taf/pkg/trustmodel/trustmodeltemplate"
	"log/slog"
)

// Holds the available functions for updating
// worker Results.
//var updateResultFuncs = map[string]ResultsUpdater{}

// Register a new ResultUpdater under a name.
// The name can be used in the config to refer to the registered function.
// The ResultUpdater is called by a worker at a point in execution when the
// Results it is responsible for should be refreshed.
/*
func RegisterUpdateResultFunc(name string, f ResultsUpdater) {
	updateResultFuncs[name] = f
}

func getUpdateResultFunc(name string) (ResultsUpdater, error) {
	if f, ok := updateResultFuncs[name]; ok {
		return f, nil
	}
	return nil, fmt.Errorf("TrustAssessmentManager: no update result function named %s registered", name)
}
*/
// later, we can make trustAssessmentManager generic, ie trustAssessmentManager[S stateT, R resultsT, M messageT]
// where stateT, resultsT and messageT are suitable interfaces.
// ToDo: make tmts fit in nicely
// ToDo: decide what is included in the state, ie workerChannels?
type trustAssessmentManager struct {
	//	mkResultsDatabase ResultsFactory
	//	updateState       StateUpdater
	//	updateResults ResultsUpdater
	tmts           map[string]trustmodeltemplate.TrustModelTemplate
	conf           config.Configuration
	workerChannels []chan core.Command
	logger         *slog.Logger
	tafContext     core.RuntimeContext
	channels       core.TafChannels
	sessions       map[string]*session.Session
	tMIs           map[string]*trustmodelinstance.TrustModelInstance
	outbox         chan core.Message
}

func NewManager(tafContext core.RuntimeContext, channels core.TafChannels) (trustAssessmentManager, error) {
	tam := trustAssessmentManager{
		//		mkResultsDatabase: func() Results { return make(map[int]int) },
		//		updateState:       updateWorkerState,
		tmts:       trustmodel.KnownTemplates,
		conf:       tafContext.Configuration,
		tafContext: tafContext,
		channels:   channels,
		sessions:   make(map[string]*session.Session),
		tMIs:       make(map[string]*trustmodelinstance.TrustModelInstance),
		logger:     logging.CreateChildLogger(tafContext.Logger, "TAM"),
	}

	var err error
	//	f, err := getUpdateResultFunc(tafContext.Configuration.TAM.UpdateResultsOp)
	if err != nil {
		return trustAssessmentManager{}, err
	}
	//	retTam.updateResults = f

	return tam, nil
}

// Get shard worker based on provided ID and configured number of shards
func (t *trustAssessmentManager) getShardWorkerById(id int) int {
	return id % t.conf.TAM.TrustModelInstanceShards
}

// Run the trust assessment trustAssessmentManager
func (t *trustAssessmentManager) Run() {

	defer func() {
		t.logger.Info("Shutting down")
	}()

	t.outbox = t.channels.OutgoingMessageChannel

	t.workerChannels = make([]chan core.Command, 0, t.conf.TAM.TrustModelInstanceShards)
	for i := range t.conf.TAM.TrustModelInstanceShards {
		ch := make(chan core.Command, 1_000)
		t.workerChannels = append(t.workerChannels, ch)

		worker := t.SpawnNewWorker(i, ch, t.tafContext)

		go worker.Run()
	}

	for {
		// Each iteration, check whether we've been cancelled.
		if err := context.Cause(t.tafContext.Context); err != nil {
			return
		}
		select {
		case <-t.tafContext.Context.Done():
			if len(t.channels.TMMChan) != 0 || len(t.channels.TAMChan) != 0 || len(t.channels.TSMChan) != 0 {
				continue
			}
			return
		case incomingCmd := <-t.channels.TAMChan:
			switch cmd := incomingCmd.(type) {
			case command.HandleRequest[tasmsg.TasInitRequest]:
				t.handleTasInitRequest(cmd)
			case command.HandleRequest[tasmsg.TasTeardownRequest]:
				t.handleTasTeardownRequest(cmd)
			default:
				t.logger.Warn("Command with no associated handling logic received by TAM", "Command Type", cmd.Type())
			}
		}
	}
}

func (t *trustAssessmentManager) createSessionId() string {

	//sessionId := fmt.Sprintf("session-%000000d", rand.IntN(999999))
	sessionId := "sessionId"

	return sessionId
}

func (t *trustAssessmentManager) handleTasInitRequest(cmd command.HandleRequest[tasmsg.TasInitRequest]) {
	t.logger.Info("Received TAS_INIT command", "Trust Model", cmd.Request.TrustModelTemplate)

	attestationCertificate, err := crypto.LoadAttestationCertificateInBase64()
	if err != nil {
		t.logger.Error("Error marshalling response", "Error", err)
	}

	//Check whether Trust Model is known
	tmt, exists := t.tmts[cmd.Request.TrustModelTemplate]
	if !exists {
		t.logger.Warn("Unknown Trust Model Template or Version:" + cmd.Request.TrustModelTemplate)

		errorMsg := "Trust model template '" + cmd.Request.TrustModelTemplate + "' could not be resolved."
		response := tasmsg.TasInitResponse{
			AttestationCertificate: attestationCertificate,
			Error:                  &errorMsg,
			SessionID:              nil,
			Success:                nil,
		}
		bytes, err := communication.BuildResponse("taf", "TAS", "TAS_INIT_RESPONSE", cmd.RequestID, response)
		if err != nil {
			t.logger.Error("Error marshalling response", "error", err)
		}
		//Send error message
		t.outbox <- core.NewMessage(bytes, "", cmd.ResponseTopic)
		return
	}
	//create session ID for client
	sessionId := t.createSessionId()
	//create Session
	newSession := session.NewInstance(sessionId, cmd.Sender)
	//put session into session map
	t.sessions[sessionId] = &newSession

	t.logger.Info("Session created:", "Session ID", newSession.ID(), "Client", newSession.Client())

	//create new TMI for session //TODO: always possible for dynamic models?
	newTMI := tmt.Spawn(cmd.Request.Params)

	//add new TMI to session
	tMIs := newSession.TrustModelInstances()
	tMIs[sessionId] = newTMI

	//add new TMI to list of all TMIs of the TAM
	t.tMIs[sessionId] = &newTMI

	t.logger.Info("TMI spawned:", "TMI ID", newTMI.ID(), "Session ID", newSession.ID(), "Client", newSession.Client())

	success := "Session with trust model template '" + newTMI.Template() + "' created."

	response := tasmsg.TasInitResponse{
		AttestationCertificate: attestationCertificate, //TODO add crypto library call
		Error:                  nil,
		SessionID:              &sessionId,
		Success:                &success,
	}

	bytes, err := communication.BuildResponse("taf", "TAS", "TAS_INIT_RESPONSE", cmd.RequestID, response)
	if err != nil {
		t.logger.Error("Error marshalling response", "error", err)
	}
	//Send response message
	t.outbox <- core.NewMessage(bytes, "", cmd.ResponseTopic)
	return
}

func (t *trustAssessmentManager) handleTasTeardownRequest(cmd command.HandleRequest[tasmsg.TasTeardownRequest]) {
	t.logger.Info("Received TAS_TEARDOWN command", "Session ID", cmd.Request.SessionID)
	_, exists := t.sessions[cmd.Request.SessionID]
	if !exists {
		errorMsg := "Session ID '" + cmd.Request.SessionID + "' not found."

		response := tasmsg.TasTeardownResponse{
			AttestationCertificate: "", //TODO add crypto library call
			Error:                  &errorMsg,
			Success:                nil,
		}
		bytes, err := communication.BuildResponse("taf", "TAS", "TAS_TEARDOWN_RESPONSE", cmd.RequestID, response)
		if err != nil {
			t.logger.Error("Error marshalling response", "error", err)
		}
		//Send error message
		t.outbox <- core.NewMessage(bytes, "", cmd.ResponseTopic)
		return
	}

	//TODO: remove session-related data

	success := "Session with ID '" + cmd.Request.SessionID + "' successfully terminated."
	response := tasmsg.TasTeardownResponse{
		AttestationCertificate: "", //TODO add crypto library call
		Error:                  nil,
		Success:                &success,
	}

	bytes, err := communication.BuildResponse("taf", "TAS", "TAS_TEARDOWN_RESPONSE", cmd.RequestID, response)
	if err != nil {
		t.logger.Error("Error marshalling response", "error", err)
	}
	//Send response message
	t.outbox <- core.NewMessage(bytes, "", cmd.ResponseTopic)
	return

}
