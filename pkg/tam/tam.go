package tam

import (
	"context"
	"fmt"
	"time"

	"gitlab-vs.informatik.uni-ulm.de/connect/taf-scalability-test/pkg/config"
	"gitlab-vs.informatik.uni-ulm.de/connect/taf-scalability-test/pkg/message"
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
	return nil, fmt.Errorf("tam: no update result function named %s registered", name)
}

// later, we can make tam generic, ie tam[S stateT, R resultsT, M messageT]
// where stateT, resultsT and messageT are suitable interfaces.
// ToDo: make tmts fit in nicely
// ToDo: decide what is included in the state, ie channels?
type tam struct {
	mkStateDatabase   StateFactory
	mkResultsDatabase ResultsFactory
	updateState       StateUpdater
	updateResults     ResultsUpdater
	tmts              TMTs
	conf              config.TAMConfiguration
}

func New(conf config.TAMConfiguration, tmts TMTs) (tam, error) {
	retTam := tam{
		mkStateDatabase:   func() State { return make(map[int][]int) },
		mkResultsDatabase: func() Results { return make(map[int]int) },
		updateState:       updateWorkerState,
		tmts:              tmts,
		conf:              conf,
	}

	var err error
	f, err := getUpdateResultFunc(conf.UpdateResultsOp)
	if err != nil {
		return tam{}, err
	}
	retTam.updateResults = f

	return retTam, nil
}

// Processes the messages received via the specified channel as fast as possible.
func (t tam) tamWorker(id int, inputs <-chan message.Message) {
	states := t.mkStateDatabase()
	results := t.mkResultsDatabase()

	// Ticker for latency benchmark
	latTicker := time.NewTicker(1 * time.Second)
	latMeasurePending := false

	for {
		select {
		case msg := <-inputs:
			t.updateState(states, t.tmts, msg)
			t.updateResults(results, states, t.tmts, msg.ID)
			time.Sleep(1 * time.Millisecond)
			if latMeasurePending && id == 0 {
				fmt.Printf("TAM: latency of %d µs\n", time.Since(msg.Timestamp).Microseconds())
				latMeasurePending = false
			}
		case <-latTicker.C:
			latMeasurePending = true
		}
	}
}

func updateWorkerState(state State, tmt TMTs, msg message.Message) {
	value, ok := tmt[msg.Type]
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

	//log.Printf("Current state for ID %d: %+v\n", msg.ID, state[msg.ID])
}

// Runs the trust assessment manager
func (t tam) Run(ctx context.Context,
	inputTMM chan message.Message,
	inputTSM chan message.Message,
	inputTAS chan message.TasQuery,
	outputTAS chan message.TasResponse) {

	defer func() {
		//log.Println("TAM: shutting down")
	}()

	// Ticker for throughput benchmark
	throughputTicker := time.NewTicker(1 * time.Second)
	lastTime := time.Now()
	msgCtr := 0

	channels := make([]chan message.Message, 0, t.conf.TrustModelInstanceShards)
	for i := range t.conf.TrustModelInstanceShards {
		ch := make(chan message.Message, 10_000)
		channels = append(channels, ch)
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
		case <-throughputTicker.C:
			delta := time.Since(lastTime)
			throughput := float64(msgCtr) / delta.Seconds()
			throughputSec := throughput
			fmt.Printf("TAM: %e messages per second\n", throughputSec)
			msgCtr = 0
			lastTime = time.Now()
		case msgFromTMM := <-inputTMM:
			//log.Printf("I am TAM, received %+v from TMM\n", msgFromTMM)
			workerId := msgFromTMM.ID % t.conf.TrustModelInstanceShards
			channels[workerId] <- msgFromTMM
			msgCtr++
		case msgFromTSM := <-inputTSM:
			//log.Printf("I am TAM, received %+v from TSM\n", msgFromTSM)
			workerId := msgFromTSM.ID % t.conf.TrustModelInstanceShards
			channels[workerId] <- msgFromTSM
			msgCtr++
			//case tasQuery := <-inputTAS:
			//	//log.Printf("I am TAM, received %+v from TAS\n", tasQuery)
			//	response := message.TasResponse{ResponseID: tasQuery.QueryID, ResponseValue: results[tasQuery.RequestedID]}
			//	outputTAS <- response

		}
	}
}
