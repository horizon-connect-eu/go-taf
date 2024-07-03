package trustassessment

import (
	"context"
	"encoding/json"
	"fmt"
	logging "github.com/vs-uulm/go-taf/internal/logger"
	"github.com/vs-uulm/go-taf/pkg/command"
	"github.com/vs-uulm/go-taf/pkg/communication"
	"github.com/vs-uulm/go-taf/pkg/config"
	"github.com/vs-uulm/go-taf/pkg/core"
	"github.com/vs-uulm/go-taf/pkg/message"
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
// ToDo: decide what is included in the state, ie channels?
type trustAssessmentManager struct {
	//	mkResultsDatabase ResultsFactory
	//	updateState       StateUpdater
	//	updateResults ResultsUpdater
	tmts       map[string]trustmodeltemplate.TrustModelTemplate
	conf       config.Configuration
	channels   []chan command.Command
	logger     *slog.Logger
	tafContext core.RuntimeContext
	sessions   map[string]*session.Session
	tMIs       map[string]*trustmodelinstance.TrustModelInstance
	outbox     chan communication.Message
}

func NewManager(tafContext core.RuntimeContext) (trustAssessmentManager, error) {
	retTam := trustAssessmentManager{
		//		mkResultsDatabase: func() Results { return make(map[int]int) },
		//		updateState:       updateWorkerState,
		tmts:       trustmodel.KnownTemplates,
		conf:       tafContext.Configuration,
		tafContext: tafContext,
		sessions:   make(map[string]*session.Session),
		tMIs:       make(map[string]*trustmodelinstance.TrustModelInstance),
	}
	retTam.logger = logging.CreateChildLogger(tafContext.Logger, "TAM")

	var err error
	//	f, err := getUpdateResultFunc(tafContext.Configuration.TAM.UpdateResultsOp)
	if err != nil {
		return trustAssessmentManager{}, err
	}
	//	retTam.updateResults = f

	return retTam, nil
}

func updateWorkerState(msg message.InternalMessage) {
	/*
		_, ok := tmt[msg.Type]
		//value, ok := tmt[msg.Type]
		if !ok {
			//log.Println("Error")
			return
		}

			_, ok = state[msg.ID]
			if !ok {
				state[msg.ID] = make([]int, 0, value+1)
			}
			state[msg.ID] = append(state[msg.ID], msg.Value)
			if len(state[msg.ID]) > value {
				state[msg.ID] = state[msg.ID][1:]
			}
	*/
	//log.Printf("Current state for ID %d: %+v\n", msg.ID, state[msg.ID])
}

// Get shard worker based on provided ID and configured number of shards
func (t *trustAssessmentManager) getShardWorkerById(id int) int {
	return id % t.conf.TAM.TrustModelInstanceShards
}

// Runs the trust assessment trustAssessmentManager
func (t *trustAssessmentManager) Run(outbox chan communication.Message) {

	defer func() {
		t.logger.Info("Shutting down")
	}()

	t.outbox = outbox

	t.channels = make([]chan command.Command, 0, t.conf.TAM.TrustModelInstanceShards)
	for i := range t.conf.TAM.TrustModelInstanceShards {
		ch := make(chan command.Command, 1_000)
		t.channels = append(t.channels, ch)

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
			/*if len(inputTMM) != 0 || len(inputTSM) != 0 {
				continue
			}*/
			return
		case incomingCmd := <-t.tafContext.TAMChan:

			switch cmd := incomingCmd.(type) {
			case command.HandleTasInitRequest:
				t.handleTasInitRequest(cmd)
			//			case command.UpdateTOCommand:
			//				t.handleUpdateTOCommand(cmd)
			case command.HandleTasTeardownRequest:
				t.handleTasTeardownRequest(cmd)
			default:
				t.logger.Warn("Unknown message received from TMM", "message", fmt.Sprintf("%+v", cmd))
			}
		}
	}
}

func (t *trustAssessmentManager) createSessionId() string {

	//sessionId := fmt.Sprintf("session-%000000d", rand.IntN(999999))
	sessionId := "sessionId"

	return sessionId
}

func (t *trustAssessmentManager) handleTasInitRequest(cmd command.HandleTasInitRequest) {
	t.logger.Info("Received TAS_INIT command", "Trust Model", cmd.Request().TrustModelTemplate)
	//Check whether Trust Model is known
	tmt, exists := t.tmts[cmd.Request().TrustModelTemplate]
	if !exists {
		t.logger.Warn("Unknown Trust Model Template or Version:" + cmd.Request().TrustModelTemplate)

		errorMsg := "Trust model template '" + cmd.Request().TrustModelTemplate + "' could not be resolved."
		response := tasmsg.TasInitResponse{
			AttestationCertificate: "", //TODO add crypto library call
			Error:                  &errorMsg,
			SessionID:              nil,
			Success:                nil,
		}
		bytes, err := buildResponse("taf", "TAS", "TAS_INIT_RESPONSE", cmd.RequestID(), response)
		if err != nil {
			t.logger.Error("Error marshalling response", "error", err)
		}
		//Send error message
		t.outbox <- communication.NewMessage(bytes, "", cmd.ResponseTopic())
		return
	}
	//create session ID for client
	sessionId := t.createSessionId()
	//create Session
	newSession := session.NewInstance(sessionId, cmd.Sender())
	//put session into session map
	t.sessions[sessionId] = &newSession

	t.logger.Info("Session created:", "Session ID", newSession.ID(), "Client", newSession.Client())

	//create new TMI for session //TODO: always possible for dynamic models?
	newTMI := tmt.Spawn(cmd.Request().Params)

	//add new TMI to session
	tMIs := newSession.TrustModelInstances()
	tMIs[sessionId] = newTMI

	//add new TMI to list of all TMIs of the TAM
	t.tMIs[sessionId] = &newTMI

	t.logger.Info("TMI spawned:", "TMI ID", newTMI.ID(), "Session ID", newSession.ID(), "Client", newSession.Client())

	success := "Session with trust model template '" + newTMI.Template() + "' created."
	response := tasmsg.TasInitResponse{
		AttestationCertificate: "", //TODO add crypto library call
		Error:                  nil,
		SessionID:              &sessionId,
		Success:                &success,
	}

	bytes, err := buildResponse("taf", "TAS", "TAS_INIT_RESPONSE", cmd.RequestID(), response)
	if err != nil {
		t.logger.Error("Error marshalling response", "error", err)
	}
	//Send response message
	t.outbox <- communication.NewMessage(bytes, "", cmd.ResponseTopic())
	return
}

func (t *trustAssessmentManager) handleTasTeardownRequest(cmd command.HandleTasTeardownRequest) {
	t.logger.Info("Received TAS_TEARDOWN command", "Session ID", cmd.Request().SessionID)
	_, exists := t.sessions[cmd.Request().SessionID]
	if !exists {
		errorMsg := "Session ID '" + cmd.Request().SessionID + "' not found."

		response := tasmsg.TasTeardownResponse{
			AttestationCertificate: "", //TODO add crypto library call
			Error:                  &errorMsg,
			Success:                nil,
		}
		bytes, err := buildResponse("taf", "TAS", "TAS_TEARDOWN_RESPONSE", cmd.RequestID(), response)
		if err != nil {
			t.logger.Error("Error marshalling response", "error", err)
		}
		//Send error message
		t.outbox <- communication.NewMessage(bytes, "", cmd.ResponseTopic())
		return
	}

	//TODO: remove session-related data

	success := "Session with ID '" + cmd.Request().SessionID + "' successfully terminated."
	response := tasmsg.TasTeardownResponse{
		AttestationCertificate: "", //TODO add crypto library call
		Error:                  nil,
		Success:                &success,
	}

	bytes, err := buildResponse("taf", "TAS", "TAS_TEARDOWN_RESPONSE", cmd.RequestID(), response)
	if err != nil {
		t.logger.Error("Error marshalling response", "error", err)
	}
	//Send response message
	t.outbox <- communication.NewMessage(bytes, "", cmd.ResponseTopic())
	return

}

type GenericResponseWrapper struct {
	Sender      string      `json:"sender"`
	ServiceType string      `json:"serviceType"`
	MessageType string      `json:"messageType"`
	ResponseId  string      `json:"responseId"`
	Message     interface{} `json:"message"`
}

/*
Function builds a byte representation of a JSON response by filling the header fields and
*/
func buildResponse(sender string, serviceType string, messageType string, responseId string, message interface{}) ([]byte, error) {
	responseWrapper := GenericResponseWrapper{
		Sender:      sender,
		ServiceType: serviceType,
		MessageType: messageType,
		ResponseId:  responseId,
		Message:     message,
	}
	return json.Marshal(responseWrapper)
}

/*
func prepareResponse() []byte {

	return
}

*/

/*
func (t *trustAssessmentManager) handleInitTMICommand(cmd command.InitTMICommand) {
	t.logger.Debug("Processing InitTMICommand", "Message", fmt.Sprintf("%+v", cmd))
	workerId := t.getShardWorkerById(int(cmd.Identifier))
	t.channels[workerId] <- cmd
}

func (t *trustAssessmentManager) handleUpdateTOCommand(cmd command.UpdateTOCommand) {
	t.logger.Debug("processing UpdateATOCommand from TMM", "Message", fmt.Sprintf("%+v", cmd))
	workerId := t.getShardWorkerById(int(cmd.Identifier))
	t.channels[workerId] <- cmd
}
*/
