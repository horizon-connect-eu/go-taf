package tam

import (
	"fmt"

	"gitlab-vs.informatik.uni-ulm.de/connect/taf-scalability-test/pkg/message"
)

func updateState(state map[int][]int, msg message.Message) {
	_, ok := state[msg.ID]
	if !ok {
		state[msg.ID] = make([]int, 0, 11)
	}
	state[msg.ID] = append(state[msg.ID], msg.Value)
	if len(state[msg.ID]) > 10 {
		state[msg.ID] = state[msg.ID][1:]
	}

	fmt.Printf("Current state for ID %d: %+v\n", msg.ID, state[msg.ID])
}

func updateResults(results map[int]int, id int, states map[int][]int) {
	sum := 0
	for _, x := range states[id] {
		sum += x
	}

	results[id] = sum
	fmt.Printf("Current sum for ID %d: %d\n", id, sum)
}

func Run(inputTMM chan message.Message, inputTSM chan message.Message) {
	states := make(map[int][]int)
	results := make(map[int]int)
	for {
		select {
		case msgFromTMM := <-inputTMM:
			fmt.Printf("I am TAM, received %+v from TMM\n", msgFromTMM)
			updateState(states, msgFromTMM)
			updateResults(results, msgFromTMM.ID, states)
		case msgFromTSM := <-inputTSM:
			fmt.Printf("I am TAM, received %+v from TSM\n", msgFromTSM)
			updateState(states, msgFromTSM)
			updateResults(results, msgFromTSM.ID, states)
		}
	}

}
