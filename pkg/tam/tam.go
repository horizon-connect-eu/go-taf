package tam

import (
	"context"
	"fmt"
	"plugin"
	"time"

	"gitlab-vs.informatik.uni-ulm.de/connect/taf-scalability-test/pkg/config"
	"gitlab-vs.informatik.uni-ulm.de/connect/taf-scalability-test/pkg/message"
)

// What are our states and results?
type State = map[int][]int
type Results = map[int]int
type TMTs = map[string]int

// later, we can make tam generic, ie tam[S stateT, R resultsT, M messageT]
// where stateT, resultsT and messageT are suitable interfaces.
// ToDo: make tmts fit in nicely
// ToDo: decide what is included in the state, ie channels?
type tam struct {
	mkStateDatabase   func() State
	mkResultsDatabase func() Results
	updateState       func(State, TMTs, message.Message)
	updateResults     func(Results, State, TMTs, int)
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
	retTam.updateResults, err = getUpdateResultsOpByName(conf.UpdateResultsOp)
	if err != nil {
		return tam{}, err
	}

	return retTam, nil
}

type X interface{}

func getUpdateResultsOpByName(name string) (func(Results, State, TMTs, int), error) {
	// TODO open questions:
	// What should be "pluginable"? only the functions or also the types? (not sure if types are even possible)
	// should we load all available plugins at startup or only the ones specified?
	// maybe load at init time using an init function would be best.
	// what are naming conventions? -> path of the .so files, names of the functions and so on.
	// Do we want to provide default functions in our own codebase?
	path := "plugins/bin/tam.so"
	p, err := plugin.Open(path)
	if err != nil {
		return nil, fmt.Errorf("could not find plugin file %s", path)
	}
	updateResultsFunc, err := p.Lookup(name)
	if err != nil {
		return nil, fmt.Errorf("could not find symbol %s in plugin file %s", name, path)
	}
	return updateResultsFunc.(func(Results, State, TMTs, int)), nil
}

// Processes the messages received via the specified channel as fast as possible.
func (t tam) tamWorker(inputs <-chan message.Message) {
	states := t.mkStateDatabase()
	results := t.mkResultsDatabase()
	for {
		msg := <-inputs
		t.updateState(states, t.tmts, msg)
		t.updateResults(results, states, t.tmts, msg.ID)
		//time.Sleep(1 * time.Millisecond)
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

	ticker := time.NewTicker(1 * time.Second)
	lastTime := time.Now()
	msgCtr := 0

	channels := make([]chan message.Message, 0, t.conf.TrustModelInstanceShards)
	for range t.conf.TrustModelInstanceShards {
		ch := make(chan message.Message, 10_000)
		channels = append(channels, ch)
		go t.tamWorker(ch)
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
		case <-ticker.C:
			delta := time.Since(lastTime)
			throughput := float64(msgCtr) / delta.Seconds()
			throughputSec := throughput
			fmt.Println("Throughput: ", throughputSec)
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
