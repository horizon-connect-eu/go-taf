package tam

import (
	"context"
	"log"

	"gitlab-vs.informatik.uni-ulm.de/connect/taf-scalability-test/pkg/message"
)

var tmt map[string]int

func updateState(state map[int][]int, msg message.Message) {
	value, ok := tmt[msg.Type]
	if !ok {
		log.Println("Error")
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

	log.Printf("Current state for ID %d: %+v\n", msg.ID, state[msg.ID])
}

// Gets the slice stored in `states` under the key `id`, computes its sum,
// and inserts this sum into `results` at key `id`.
func updateResults(results map[int]int, id int, states map[int][]int) {
	sum := 0
	for _, x := range states[id] {
		sum += x
	}

	results[id] = sum
	log.Printf("Current sum for ID %d: %d\n", id, sum)
}

// Runs the trust assessment manager
func Run(ctx context.Context, tmts map[string]int, inputTMM chan message.Message, inputTSM chan message.Message, inputTAS chan message.TasQuery, outputTAS chan message.TasResponse) {
	defer func() {
		log.Println("TAM: shutting down")
	}()

	tmt = tmts

	states := make(map[int][]int)
	results := make(map[int]int)
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
		case msgFromTMM := <-inputTMM:
			log.Printf("I am TAM, received %+v from TMM\n", msgFromTMM)
			updateState(states, msgFromTMM)
			updateResults(results, msgFromTMM.ID, states)
		case msgFromTSM := <-inputTSM:
			log.Printf("I am TAM, received %+v from TSM\n", msgFromTSM)
			updateState(states, msgFromTSM)
			updateResults(results, msgFromTSM.ID, states)
		case tasQuery := <-inputTAS:
			log.Printf("I am TAM, received %+v from TAS\n", tasQuery)
			response := message.TasResponse{ResponseID: tasQuery.QueryID, ResponseValue: results[tasQuery.RequestedID]}
			outputTAS <- response

		}
	}
}
