package trustassessment

import (
	"context"
	"fmt"
	logging "github.com/vs-uulm/go-taf/internal/logger"
	"github.com/vs-uulm/go-taf/pkg/config"
	"github.com/vs-uulm/go-taf/pkg/core"
	"github.com/vs-uulm/go-taf/pkg/message"
	"github.com/vs-uulm/go-taf/pkg/trustmodel/trustmodelinstance"
	"log/slog"
)

// Holds the available functions for updating
// worker Results.
var updateResultFuncs = map[string]ResultsUpdater{}

// Register a new ResultUpdater under a name.
// The name can be used in the config to refer to the registered function.
// The ResultUpdater is called by a worker at a point in execution when the
// Results it is responsible for should be refreshed.
func RegisterUpdateResultFunc(name string, f ResultsUpdater) {
	updateResultFuncs[name] = f
}

func getUpdateResultFunc(name string) (ResultsUpdater, error) {
	if f, ok := updateResultFuncs[name]; ok {
		return f, nil
	}
	return nil, fmt.Errorf("TrustAssessmentManager: no update result function named %s registered", name)
}

// later, we can make trustAssessmentManager generic, ie trustAssessmentManager[S stateT, R resultsT, M messageT]
// where stateT, resultsT and messageT are suitable interfaces.
// ToDo: make tmts fit in nicely
// ToDo: decide what is included in the state, ie channels?
type trustAssessmentManager struct {
	mkStateDatabase   StateFactory
	mkResultsDatabase ResultsFactory
	updateState       StateUpdater
	updateResults     ResultsUpdater
	tmts              TMTs
	conf              config.Configuration
	channels          []chan Command
	logger            *slog.Logger
	tafContext        core.RuntimeContext
}

func NewManager(tafContext core.RuntimeContext, tmts TMTs) (trustAssessmentManager, error) {
	retTam := trustAssessmentManager{
		mkStateDatabase:   func() State { return make(map[int]trustmodelinstance.TrustModelInstance) },
		mkResultsDatabase: func() Results { return make(map[int]int) },
		updateState:       updateWorkerState,
		tmts:              tmts,
		conf:              tafContext.Configuration,
		tafContext:        tafContext,
	}
	retTam.logger = logging.CreateChildLogger(tafContext.Logger, "TAM")

	var err error
	f, err := getUpdateResultFunc(tafContext.Configuration.TAM.UpdateResultsOp)
	if err != nil {
		return trustAssessmentManager{}, err
	}
	retTam.updateResults = f

	return retTam, nil
}

func updateWorkerState(state State, tmt TMTs, msg message.InternalMessage) {
	_, ok := tmt[msg.Type]
	//value, ok := tmt[msg.Type]
	if !ok {
		//log.Println("Error")
		return
	}

	/*
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
func (t *trustAssessmentManager) Run(
	inputTMM chan Command,
	inputTSM chan Command) {

	defer func() {
		t.logger.Info("Shutting down")
	}()

	t.channels = make([]chan Command, 0, t.conf.TAM.TrustModelInstanceShards)
	for i := range t.conf.TAM.TrustModelInstanceShards {
		ch := make(chan Command, 1_000)
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
		case cmdFromTMM := <-inputTMM:

			switch cmd := cmdFromTMM.(type) {
			case InitTMICommand:
				t.handleInitTMICommand(cmd)
			default:
				t.logger.Warn("Unknown message received from TMM", "message", fmt.Sprintf("%+v", cmd))
			}
		case cmdFromTSM := <-inputTSM:

			switch cmd := cmdFromTSM.(type) {
			case UpdateTOCommand:
				t.handleUpdateTOCommand(cmd)
			default:
				t.logger.Warn("Unknown message received from TMM", "message", fmt.Sprintf("%+v", cmd))
			}
		}
	}
}

func (t *trustAssessmentManager) handleInitTMICommand(cmd InitTMICommand) {
	t.logger.Debug("Processing InitTMICommand", "Message", fmt.Sprintf("%+v", cmd))
	workerId := t.getShardWorkerById(int(cmd.Identifier))
	t.channels[workerId] <- cmd
}

func (t *trustAssessmentManager) handleUpdateTOCommand(cmd UpdateTOCommand) {
	t.logger.Debug("processing UpdateATOCommand from TMM", "Message", fmt.Sprintf("%+v", cmd))
	workerId := t.getShardWorkerById(int(cmd.Identifier))
	t.channels[workerId] <- cmd
}
