package trustassessment

import (
	"time"

	"github.com/vs-uulm/go-taf/pkg/trustmodel/instance"
)

// Processes the messages received via the specified channel as fast as possible.
func (t *trustAssessmentManager) tamWorker(id int, inputs <-chan Command) {
	states := t.mkStateDatabase()
	//results := t.mkResultsDatabase()

	// Ticker for latency benchmark
	latTicker := time.NewTicker(1 * time.Second)
	latMeasurePending := false

	for {
		select {
		case command := <-inputs:
			processCommand(id, command, states)
			if latMeasurePending && id == 0 {
				//fmt.Printf("TAM: latency of %d µs\n", time.Since(command.Timestamp).Microseconds())
				latMeasurePending = false
			}
		case <-latTicker.C:
			latMeasurePending = true

		}
	}
}

func processCommand(workerID int, cmd Command, states State) {
	switch cmd := cmd.(type) {
	case InitTMICommand:
		//LOG: fmt.Printf("[TAM Worker %d] handling InitTMICommand: %v\n", workerID, cmd)

		states[int(cmd.Identifier)] = instance.NewTrustModelInstance(int(cmd.Identifier), cmd.TrustModelTemplate)

	case UpdateTOCommand:
		//LOG: fmt.Printf("[TAM Worker %d] handling UpdateATOCommand: %v\n", workerID, cmd)

		//trustModelInstance := states[int(cmd.Identifier)]

		//LOG: fmt.Printf("[TAM Worker %d] updating TMI %d\n", workerID, trustModelInstance.GetId())

	default:
		//LOG: fmt.Printf("[TAM Worker %d] Unknown message to %v\n", workerID, cmd)
	}
}
