package trustassessment

import (
	"context"
	"fmt"
	"gitlab-vs.informatik.uni-ulm.de/connect/taf-brussels-demo/pkg/config"
	"gitlab-vs.informatik.uni-ulm.de/connect/taf-brussels-demo/pkg/message"
	"gitlab-vs.informatik.uni-ulm.de/connect/taf-brussels-demo/pkg/trustmodel/instance"
	"log"
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
}

func NewManager(conf config.Configuration, tmts TMTs) (trustAssessmentManager, error) {
	retTam := trustAssessmentManager{
		mkStateDatabase:   func() State { return make(map[int]instance.TrustModelInstance) },
		mkResultsDatabase: func() Results { return make(map[int]int) },
		updateState:       updateWorkerState,
		tmts:              tmts,
		conf:              conf,
	}

	var err error
	f, err := getUpdateResultFunc(conf.TAM.UpdateResultsOp)
	if err != nil {
		return trustAssessmentManager{}, err
	}
	retTam.updateResults = f

	return retTam, nil
}

// Processes the messages received via the specified channel as fast as possible.
func (t *trustAssessmentManager) tamWorker(id int, inputs <-chan Command) {
	states := t.mkStateDatabase()
	//results := t.mkResultsDatabase()

	// Ticker for latency benchmark
	//latTicker := time.NewTicker(1 * time.Second)
	//latMeasurePending := false

	for {
		select {
		case command := <-inputs:
			switch cmd := command.(type) {
			case InitTMICommand:
				fmt.Printf("[TAM Worker %d] handling InitTMICommand: %v", id, cmd)

				states[int(cmd.Identifier)] = instance.NewTrustModelInstance(int(cmd.Identifier), cmd.TrustModelTemplate)

			case UpdateATOCommand:
				fmt.Printf("[TAM Worker %d] handling UpdateATOCommand: %v", id, cmd)

				trustModelInstance := states[int(cmd.Identifier)]
				println(trustModelInstance.GetId())

			default:
				fmt.Printf("[TAM Worker %d] Unknown message to %v", id, cmd)
			}
			/*
				t.updateState(states, t.tmts, msg)
				t.updateResults(results, states, t.tmts, msg.ID)
				//time.Sleep(1 * time.Millisecond)
				if latMeasurePending && id == 0 {
					fmt.Printf("TAM: latency of %d µs\n", time.Since(msg.Timestamp).Microseconds())
					latMeasurePending = false
				}

			*/
			/*
				case <-latTicker.C:
						latMeasurePending = true
			*/
		}
	}
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
func (t *trustAssessmentManager) Run(ctx context.Context,
	inputTMM chan Command,
	inputTSM chan Command) {

	defer func() {
		//log.Println("TAM: shutting down")
	}()

	t.channels = make([]chan Command, 0, t.conf.TAM.TrustModelInstanceShards)
	for i := range t.conf.TAM.TrustModelInstanceShards {
		ch := make(chan Command, 1_000)
		t.channels = append(t.channels, ch)
		go t.tamWorker(i, ch)
	}

	for {
		// Each iteration, check whether we've been cancelled.
		if err := context.Cause(ctx); err != nil {
			return
		}
		select {
		case <-ctx.Done():
			/*if len(inputTMM) != 0 || len(inputTSM) != 0 {
				continue
			}*/
			return
		case cmdFromTMM := <-inputTMM:

			switch cmd := cmdFromTMM.(type) {
			case InitTMICommand:
				t.handleInitTMICommand(cmd)
			default:
				log.Printf("[TAM] Unknown message %+v from TMM\n", cmd)
			}
		case cmdFromTSM := <-inputTSM:

			switch cmd := cmdFromTSM.(type) {
			case UpdateATOCommand:
				t.handleUpdateATOCommand(cmd)
			default:
				log.Printf("[TAM] Unknown message %+v from TMM\n", cmd)
			}
		}
	}
}

func (t *trustAssessmentManager) handleInitTMICommand(cmd InitTMICommand) {
	//	log.Printf("[TAM] processing InitTMICommand %+v from TMM\n", cmd)
	workerId := t.getShardWorkerById(int(cmd.Identifier))
	t.channels[workerId] <- cmd
}

func (t *trustAssessmentManager) handleUpdateATOCommand(cmd UpdateATOCommand) {
	//	log.Printf("[TAM] processing UpdateATOCommand %+v from TMM\n", cmd)
	workerId := t.getShardWorkerById(int(cmd.Identifier))
	t.channels[workerId] <- cmd
}
