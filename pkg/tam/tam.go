package tam

import (
	"context"
	"fmt"
	"time"

	"gitlab-vs.informatik.uni-ulm.de/connect/taf-scalability-test/pkg/config"
	"gitlab-vs.informatik.uni-ulm.de/connect/taf-scalability-test/pkg/message"
)

var tmt map[string]int

func updateState(state map[int][]int, msg message.Message) {
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

// Gets the slice stored in `states` under the key `id`, computes its sum,
// and inserts this sum into `results` at key `id`.
func updateResults(results map[int]int, id int, states map[int][]int) {
	sum := 0
	for _, x := range states[id] {
		sum += x
	}

	results[id] = sum
	//log.Printf("Current sum for ID %d: %d\n", id, sum)
}

func worker(inputs <-chan message.Message) {
	states := make(map[int][]int)
	results := make(map[int]int)
	for {
		msg := <-inputs
		updateState(states, msg)
		updateResults(results, msg.ID, states)
		//time.Sleep(1 * time.Millisecond)
	}
}

// Runs the trust assessment manager
func Run(ctx context.Context, tmts map[string]int, tamConfig config.TAMConfiguration, inputTMM chan message.Message, inputTSM chan message.Message, inputTAS chan message.TasQuery, outputTAS chan message.TasResponse) {
	defer func() {
		//log.Println("TAM: shutting down")
	}()

	tmt = tmts

	//states := make(map[int][]int)
	//results := make(map[int]int)

	ticker := time.NewTicker(1 * time.Second)
	lastTime := time.Now()
	msgCtr := 0

	channels := make([]chan message.Message, 0, tamConfig.TrustModelInstanceShards)
	for range tamConfig.TrustModelInstanceShards {
		ch := make(chan message.Message, 10_000)
		channels = append(channels, ch)
		go worker(ch)
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
			workerId := msgFromTMM.ID % tamConfig.TrustModelInstanceShards
			channels[workerId] <- msgFromTMM
			msgCtr++
		case msgFromTSM := <-inputTSM:
			//log.Printf("I am TAM, received %+v from TSM\n", msgFromTSM)
			workerId := msgFromTSM.ID % tamConfig.TrustModelInstanceShards
			channels[workerId] <- msgFromTSM
			msgCtr++
			//case tasQuery := <-inputTAS:
			//	//log.Printf("I am TAM, received %+v from TAS\n", tasQuery)
			//	response := message.TasResponse{ResponseID: tasQuery.QueryID, ResponseValue: results[tasQuery.RequestedID]}
			//	outputTAS <- response

		}
	}
}
