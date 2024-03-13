package tam

import (
	"context"
	"fmt"
	"time"

	"gitlab-vs.informatik.uni-ulm.de/connect/taf-scalability-test/pkg/config"
	"gitlab-vs.informatik.uni-ulm.de/connect/taf-scalability-test/pkg/message"
)

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

// Gets the slice stored in `states` under the key `id`, computes its sum,
// and inserts this sum into `results` at key `id`.
func updateWorkerResultsAdd(results Results, states State, tmts TMTs, id int) {
	sum := 0
	for _, x := range states[id] {
		sum += x
	}
	results[id] = sum
	//log.Printf("Current sum for ID %d: %d\n", id, sum)
}

// Gets the slice stored in `states` under the key `id`, computes its product,
// and inserts this sum into `results` at key `id`.
func updateWorkerResultsMult(results Results, states State, tmts TMTs, id int) {
	prod := 1
	for _, x := range states[id] {
		prod *= x
	}
	results[id] = prod
	//log.Printf("Current sum for ID %d: %d\n", id, sum)
}

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

func NewDefault(conf config.TAMConfiguration, tmts TMTs) (tam, error) {
	retTam := tam{
		mkStateDatabase:   func() State { return make(map[int][]int) },
		mkResultsDatabase: func() Results { return make(map[int]int) },
		updateState:       updateWorkerState,
		updateResults:     updateWorkerResultsAdd,
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

func (t *tam) SetUpdateResults(f func(Results, State, TMTs, int)) {
	t.updateResults = f
}

func (t *tam) SetUpdateState(f func(State, TMTs, message.Message)) {
	t.updateState = f
}

func getUpdateResultsOpByName(name string) (func(Results, State, TMTs, int), error) {
	switch name {
	case "Add":
		return updateWorkerResultsAdd, nil
	case "Mult":
		return updateWorkerResultsMult, nil
	default:
		return nil, fmt.Errorf("tam: no update results function of name \"%s\"", name)
	}
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
