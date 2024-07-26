package trustassessment

import (
	"context"
	"crypto-library-interface/pkg/crypto"
	"fmt"
	logging "github.com/vs-uulm/go-taf/internal/logger"
	"github.com/vs-uulm/go-taf/pkg/command"
	"github.com/vs-uulm/go-taf/pkg/communication"
	"github.com/vs-uulm/go-taf/pkg/config"
	"github.com/vs-uulm/go-taf/pkg/core"
	"github.com/vs-uulm/go-taf/pkg/manager"
	messages "github.com/vs-uulm/go-taf/pkg/message"
	aivmsg "github.com/vs-uulm/go-taf/pkg/message/aiv"
	mbdmsg "github.com/vs-uulm/go-taf/pkg/message/mbd"
	tasmsg "github.com/vs-uulm/go-taf/pkg/message/tas"
	v2xmsg "github.com/vs-uulm/go-taf/pkg/message/v2x"
	"github.com/vs-uulm/go-taf/pkg/trustmodel/session"
	"log/slog"
)

type Manager struct {
	conf           config.Configuration
	workerChannels []chan core.Command
	logger         *slog.Logger
	tafContext     core.TafContext
	channels       core.TafChannels
	sessions       map[string]*session.Session
	tMIs           map[string]*core.TrustModelInstance
	outbox         chan core.Message
	tsm            manager.TrustSourceManager
	tmm            manager.TrustModelManager
}

func NewManager(tafContext core.TafContext, channels core.TafChannels) (*Manager, error) {
	tam := &Manager{
		//		mkResultsDatabase: func() Results { return make(map[int]int) },
		//		updateState:       updateWorkerState,
		conf:       tafContext.Configuration,
		tafContext: tafContext,
		channels:   channels,
		sessions:   make(map[string]*session.Session),
		tMIs:       make(map[string]*core.TrustModelInstance),
		logger:     logging.CreateChildLogger(tafContext.Logger, "TAM"),
	}
	tam.logger.Info("Initializing Trust Assessment Manager", "Worker Count", tam.conf.TAM.TrustModelInstanceShards)
	return tam, nil
}

func (tam *Manager) SetManagers(managers manager.TafManagers) {
	tam.tmm = managers.TMM
	tam.tsm = managers.TSM
}

// Get shard worker based on provided ID and configured number of shards
func (tam *Manager) getShardWorkerById(id int) int {
	return id % tam.conf.TAM.TrustModelInstanceShards
}

// Run the trust assessment manager
func (tam *Manager) Run() {

	defer func() {
		tam.logger.Info("Shutting down")
	}()

	tam.outbox = tam.channels.OutgoingMessageChannel
	tsm := tam.tsm
	tmm := tam.tmm

	tam.workerChannels = make([]chan core.Command, 0, tam.conf.TAM.TrustModelInstanceShards)
	for i := range tam.conf.TAM.TrustModelInstanceShards {
		ch := make(chan core.Command, 1_000)
		tam.workerChannels = append(tam.workerChannels, ch)

		worker := tam.SpawnNewWorker(i, ch, tam.tafContext)

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
		case incomingCmd := <-tam.channels.TAMChannel:
			switch cmd := incomingCmd.(type) {
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
			case command.HandleRequest[tasmsg.TasInitRequest]:
				tam.HandleTasInitRequest(cmd)
			case command.HandleRequest[tasmsg.TasTeardownRequest]:
				tam.HandleTasTeardownRequest(cmd)
			case command.HandleOneWay[v2xmsg.V2XCpm]:
				tmm.HandleV2xCpmMessage(cmd)
			default:
				tam.logger.Warn("Command with no associated handling logic received by TAM", "Command Type", cmd.Type())
			}
		}
	}
}

func (tam *Manager) Hello() {
	fmt.Println("hello")
}

func (tam *Manager) createSessionId() string {

	//sessionId := fmt.Sprintf("session-%000000d", rand.IntN(999999))
	sessionId := "sessionId"

	return sessionId
}

func (tam *Manager) HandleTasInitRequest(cmd command.HandleRequest[tasmsg.TasInitRequest]) {
	tam.logger.Info("Received TAS_INIT command", "Trust Model", cmd.Request.TrustModelTemplate)

	attestationCertificate, err := crypto.LoadAttestationCertificateInBase64()
	if err != nil {
		tam.logger.Error("Error marshalling response", "Error", err)
	}

	tmt := tam.tmm.ResolveTMT(cmd.Request.TrustModelTemplate)
	if tmt == nil {
		tam.logger.Warn("Unknown Trust Model Template or Version:" + cmd.Request.TrustModelTemplate)

		errorMsg := "Trust model template '" + cmd.Request.TrustModelTemplate + "' could not be resolved."
		response := tasmsg.TasInitResponse{
			AttestationCertificate: attestationCertificate,
			Error:                  &errorMsg,
			SessionID:              nil,
			Success:                nil,
		}
		bytes, err := communication.BuildResponse(tam.tafContext.Configuration.Communication.TafEndpoint, messages.TAS_INIT_RESPONSE, cmd.RequestID, response)
		if err != nil {
			tam.logger.Error("Error marshalling response", "error", err)
		}
		//Send error message
		tam.outbox <- core.NewMessage(bytes, "", cmd.ResponseTopic)
		return
	}
	tmt.EvidenceSources()
	//create session ID for client
	sessionId := tam.createSessionId()
	//create Session
	newSession := session.NewInstance(sessionId, cmd.Sender)
	//put session into session map
	tam.sessions[sessionId] = &newSession

	tam.logger.Info("Session created:", "Session ID", newSession.ID(), "Client", newSession.Client())

	//create new TMI for session //TODO: always possible for dynamic models?
	newTMI := tmt.Spawn(cmd.Request.Params, tam.tafContext, tam.channels)
	//add new TMI to session
	tMIs := newSession.TrustModelInstances()
	tMIs[sessionId] = newTMI

	//add new TMI to list of all TMIs of the TAM
	tam.tMIs[sessionId] = &newTMI
	tam.logger.Info("TMI spawned:", "TMI ID", newTMI.ID(), "Session ID", newSession.ID(), "Client", newSession.Client())

	//Initialize TMI
	newTMI.Init()

	tam.tsm.InitTrustSourceQuantifiers(newTMI)

	success := "Session with trust model template '" + newTMI.Template().TemplateName() + "@" + newTMI.Template().Version() + "' created."

	response := tasmsg.TasInitResponse{
		AttestationCertificate: attestationCertificate, //TODO add crypto library call
		Error:                  nil,
		SessionID:              &sessionId,
		Success:                &success,
	}

	bytes, err := communication.BuildResponse("taf", messages.TAS_INIT_RESPONSE, cmd.RequestID, response)
	if err != nil {
		tam.logger.Error("Error marshalling response", "error", err)
	}
	//Send response message
	tam.outbox <- core.NewMessage(bytes, "", cmd.ResponseTopic)
	return
}

func (tam *Manager) HandleTasTeardownRequest(cmd command.HandleRequest[tasmsg.TasTeardownRequest]) {
	tam.logger.Info("Received TAS_TEARDOWN command", "Session ID", cmd.Request.SessionID)
	_, exists := tam.sessions[cmd.Request.SessionID]
	if !exists {
		errorMsg := "Session ID '" + cmd.Request.SessionID + "' not found."

		response := tasmsg.TasTeardownResponse{
			AttestationCertificate: "", //TODO add crypto library call
			Error:                  &errorMsg,
			Success:                nil,
		}
		bytes, err := communication.BuildResponse("taf", messages.TAS_TEARDOWN_RESPONSE, cmd.RequestID, response)
		if err != nil {
			tam.logger.Error("Error marshalling response", "error", err)
		}
		//Send error message
		tam.outbox <- core.NewMessage(bytes, "", cmd.ResponseTopic)
		return
	}

	//TODO: remove session-related data

	success := "Session with ID '" + cmd.Request.SessionID + "' successfully terminated."
	response := tasmsg.TasTeardownResponse{
		AttestationCertificate: "", //TODO add crypto library call
		Error:                  nil,
		Success:                &success,
	}

	bytes, err := communication.BuildResponse("taf", messages.TAS_TEARDOWN_RESPONSE, cmd.RequestID, response)
	if err != nil {
		tam.logger.Error("Error marshalling response", "error", err)
	}
	//Send response message
	tam.outbox <- core.NewMessage(bytes, "", cmd.ResponseTopic)
	return

}
