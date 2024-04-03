package trustassessment

import (
	"fmt"
	"time"

	"github.com/vs-uulm/go-taf/pkg/trustmodel/instance"
	"github.com/vs-uulm/taf-tlee-interface/pkg/subjectivelogic"
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

		var evidence map[string]bool
		var omega_DTI subjectivelogic.Opinion
		var omega *subjectivelogic.Opinion

		if cmd.Trustee == "1" {
			evidence = states[int(cmd.Identifier)].Evidence1
			omega_DTI = states[int(cmd.Identifier)].Omega_DTI_1
		} else if cmd.Trustee == "2" {
			evidence = states[int(cmd.Identifier)].Evidence2
			omega_DTI = states[int(cmd.Identifier)].Omega_DTI_2
		} else {
			return
		}

		evidence[cmd.TS_ID] = cmd.Evidence

		omega = &omega_DTI
		omega.Belief = omega.Belief + 0.1

		fmt.Println("Omega: ", omega)
		fmt.Println("DTI:", omega_DTI)

	default:
		//LOG: fmt.Printf("[TAM Worker %d] Unknown message to %v\n", workerID, cmd)
	}
}
