package trustassessment

import (
	"context"
	"fmt"
	"github.com/vs-uulm/go-taf/internal/logger"
	"github.com/vs-uulm/go-taf/pkg/command"
	"github.com/vs-uulm/go-taf/pkg/core"
	"github.com/vs-uulm/go-taf/pkg/trustdecision"
	"github.com/vs-uulm/taf-tlee-interface/pkg/tleeinterface"
	"log/slog"
)

type Worker struct {
	tafContext   core.TafContext
	id           int
	workerQueue  <-chan core.Command
	logger       *slog.Logger
	tmis         map[string]core.TrustModelInstance
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
	worker.logger.Info("Deleting Trust Model Instance with ID " + cmd.TmiID)
	tmi, exists := worker.tmis[cmd.TmiID]
	if !exists {
		return
	}
	tmi.Cleanup()
	delete(worker.tmis, cmd.TmiID)
	delete(worker.tmiSessions, cmd.TmiID)
}

func (worker *Worker) handleTMIInit(cmd command.HandleTMIInit) {
	worker.logger.Info("Registering new Trust Model Instance with ID " + cmd.TmiID)
	worker.tmis[cmd.TmiID] = cmd.TMI
	worker.tmiSessions[cmd.TmiID] = cmd.SessionID
}

func (worker *Worker) handleTMIUpdate(cmd command.HandleTMIUpdate) {
	worker.logger.Info("Updating Trust Model Instance with ID " + cmd.TmiID)
	tmi, exists := worker.tmis[cmd.TmiID]
	if !exists {
		return
	}
	sessionID, _ := worker.tmiSessions[cmd.TmiID]

	tmi.Update(cmd.Update)
	atls := worker.tlee.RunTLEE(tmi.ID(), tmi.Version(), tmi.Fingerprint(), tmi.Structure(), tmi.Values())
	worker.logger.Warn("TLEE called", "Results", fmt.Sprintf("%+v", atls))
	projectedProbabilities := make(map[string]float64, len(atls))
	trustDecisions := make(map[string]bool, len(atls))
	for proposition, opinion := range atls {
		projectedProbabilities[proposition] = trustdecision.ProjectProbability(opinion)
		//trustDecisions[proposition] = trustdecision.Decide(opinion, ) TODO: How to get RTL?
	}
	resultSet := core.CreateAtlResultSet(tmi.ID(), sessionID, tmi.Version(), atls, projectedProbabilities, trustDecisions)
	atlUpdateCmd := command.CreateHandleATLUpdate(resultSet)
	worker.workersToTam <- atlUpdateCmd
}
