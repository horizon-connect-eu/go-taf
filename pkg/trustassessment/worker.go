package trustassessment

import (
	"context"
	"fmt"
	"github.com/vs-uulm/go-subjectivelogic/pkg/subjectivelogic"
	"github.com/vs-uulm/go-taf/internal/logger"
	"github.com/vs-uulm/go-taf/pkg/command"
	"github.com/vs-uulm/go-taf/pkg/core"
	"github.com/vs-uulm/go-taf/pkg/trustdecision"
	"github.com/vs-uulm/taf-tlee-interface/pkg/tleeinterface"
	"log/slog"
)

type Worker struct {
	tafContext  core.TafContext
	id          int
	workerQueue <-chan core.Command
	logger      *slog.Logger
	//full tmiID->TMI
	tmis map[string]core.TrustModelInstance
	//tmiID->SessionID
	tmiSessions  map[string]string
	workersToTam chan<- core.Command
	tlee         tleeinterface.TLEE
}

func (tam *Manager) SpawnNewWorker(id int, workerQueue <-chan core.Command, workersToTam chan<- core.Command, tafContext core.TafContext, tlee tleeinterface.TLEE) Worker {
	return Worker{
		tafContext:   tafContext,
		id:           id,
		workerQueue:  workerQueue,
		logger:       logger.CreateChildLogger(tafContext.Logger, fmt.Sprintf("TAM-WORKER-%d", id)),
		tmis:         make(map[string]core.TrustModelInstance),
		tmiSessions:  make(map[string]string),
		workersToTam: workersToTam,
		tlee:         tlee,
	}
}

func (worker *Worker) Run() {
	defer func() {
		worker.logger.Info("Shutting down")
	}()

	for {
		// Each iteration, check whether we've been cancelled.
		if err := context.Cause(worker.tafContext.Context); err != nil {
			return
		}
		select {
		case <-worker.tafContext.Context.Done():
			if len(worker.workerQueue) != 0 {
				continue
			}
			return
		case incomingCmd := <-worker.workerQueue:
			switch cmd := incomingCmd.(type) {
			case command.HandleTMIUpdate:
				worker.handleTMIUpdate(cmd)
			case command.HandleTMIInit:
				worker.handleTMIInit(cmd)
			case command.HandleTMIDestroy:
				worker.handleTMIDestroy(cmd)
			default:
				worker.logger.Warn("Command with no associated handling logic received by Worker", "Command Type", cmd.Type())
			}
		}
	}
}

func (worker *Worker) handleTMIDestroy(cmd command.HandleTMIDestroy) {
	worker.logger.Info("Deleting Trust Model Instance with ID " + cmd.FullTMI)
	tmi, exists := worker.tmis[cmd.FullTMI]
	if !exists {
		return
	}
	tmi.Cleanup()
	delete(worker.tmis, cmd.FullTMI)
	delete(worker.tmiSessions, cmd.FullTMI)
}

func (worker *Worker) handleTMIInit(cmd command.HandleTMIInit) {
	worker.logger.Info("Registering new Trust Model Instance with ID " + cmd.FullTMI)
	worker.tmis[cmd.FullTMI] = cmd.TMI
	_, session, _, _ := core.SplitFullTMIIdentifier(cmd.FullTMI)
	worker.tmiSessions[cmd.FullTMI] = session

	//Run TLEE
	atls := worker.executeTLEE(worker.tmis[cmd.FullTMI])
	//Run TDE
	resultSet := worker.executeTDE(worker.tmis[cmd.FullTMI], atls)

	atlUpdateCmd := command.CreateHandleATLUpdate(resultSet, cmd.FullTMI)
	worker.workersToTam <- atlUpdateCmd
}

func (worker *Worker) executeTLEE(tmi core.TrustModelInstance) map[string]subjectivelogic.QueryableOpinion {
	var atls map[string]subjectivelogic.QueryableOpinion
	//Only call TLEE when the graph structure is existing and not empty; otherwise skip and return empty ATL set
	if tmi.Structure() != nil && len(tmi.Structure().AdjacencyList()) > 0 { //TODO: Values?
		worker.logger.Debug("TLEE Input", "TMI", tmi.String())
		atls = worker.tlee.RunTLEE(tmi.ID(), tmi.Version(), tmi.Fingerprint(), tmi.Structure(), tmi.Values())
		worker.logger.Debug("TLEE called", "Results", fmt.Sprintf("%+v", atls))
	} else {
		atls = make(map[string]subjectivelogic.QueryableOpinion)
		worker.logger.Debug("TLEE call omitted due to empty TMI", "Results", fmt.Sprintf("%+v", atls))
	}
	return atls
}

func (worker *Worker) executeTDE(tmi core.TrustModelInstance, atls map[string]subjectivelogic.QueryableOpinion) core.AtlResultSet {
	rtls := tmi.RTLs()
	projectedProbabilities := make(map[string]float64, len(atls))
	trustDecisions := make(map[string]core.TrustDecision, len(atls))
	for proposition, atlOpinion := range atls {
		rtlOpinion, exists := rtls[proposition]
		if !exists {
			worker.logger.Error("Could not find RTL in trust model instance for proposition "+proposition, "TMI ID", tmi.ID())
			trustDecisions[proposition] = core.UNDECIDABLE //If no RTL is found, we set trust decision to UNDECIDABLE as default
		} else {
			trustDecisions[proposition] = trustdecision.Decide(atlOpinion, rtlOpinion)
		}
		projectedProbabilities[proposition] = trustdecision.ProjectProbability(atlOpinion)
	}
	resultSet := core.CreateAtlResultSet(tmi.ID(), tmi.Version(), atls, projectedProbabilities, trustDecisions)
	return resultSet
}

func (worker *Worker) handleTMIUpdate(cmd command.HandleTMIUpdate) {
	worker.logger.Info("Updating Trust Model Instance with ID " + cmd.FullTmiID)

	tmi, exists := worker.tmis[cmd.FullTmiID]
	if !exists {
		return
	}
	//sessionID, _ := worker.tmiSessions[cmd.TmiID]

	//Execute TMI Updates
	for _, update := range cmd.Updates {
		tmi.Update(update)
	}
	//Run TLEE
	atls := worker.executeTLEE(tmi)
	//Run TDE
	resultSet := worker.executeTDE(tmi, atls)

	atlUpdateCmd := command.CreateHandleATLUpdate(resultSet, cmd.FullTmiID)
	worker.workersToTam <- atlUpdateCmd
}
