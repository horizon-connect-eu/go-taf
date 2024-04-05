package trustassessment

import (
	"fmt"
	"math"
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

		var evidence_collection map[string]bool
		var omega_DTI subjectivelogic.Opinion
		var omega subjectivelogic.Opinion

		if cmd.Trustee == "1" {
			evidence_collection = states[int(cmd.Identifier)].Evidence1
			omega_DTI = states[int(cmd.Identifier)].Omega_DTI_1
		} else if cmd.Trustee == "2" {
			evidence_collection = states[int(cmd.Identifier)].Evidence2
			omega_DTI = states[int(cmd.Identifier)].Omega_DTI_2
		} else {
			return
		}

		evidence_collection[cmd.TS_ID] = cmd.Evidence
		omega = omega_DTI

		for ts_id, evidence := range evidence_collection {
			// Equation: delta = u_DTI * weight_ts -> delta specifies how much belief, disbelief and uncertainty will be increased / decreased
			if evidence { // positive evidence, e.g. secure boot ran successfully
				omega.Belief = omega.Belief + omega_DTI.Uncertainty*states[int(cmd.Identifier)].Weights[ts_id]
				omega.Uncertainty = omega.Uncertainty - omega_DTI.Uncertainty*states[int(cmd.Identifier)].Weights[ts_id]
			} else if !evidence { // negative evidence, e.g. secure boot didn't run successfully
				omega.Disbelief = omega.Disbelief + omega_DTI.Uncertainty*states[int(cmd.Identifier)].Weights[ts_id]
				omega.Uncertainty = omega.Uncertainty - omega_DTI.Uncertainty*states[int(cmd.Identifier)].Weights[ts_id]
			}
		}

		if entry, ok := states[int(cmd.Identifier)]; ok {
			// round values to two decimal places
			omega.Belief = math.Abs(math.Round(omega.Belief*100) / 100)
			omega.Disbelief = math.Abs(math.Round(omega.Disbelief*100) / 100)
			omega.Uncertainty = math.Abs(math.Round(omega.Uncertainty*100) / 100)

			if cmd.Trustee == "1" {
				entry.Omega1 = omega
			} else if cmd.Trustee == "2" {
				entry.Omega2 = omega
			}

			states[int(cmd.Identifier)] = entry

			if cmd.Trustee == "1" {
				fmt.Println("Omega1 :", omega)
				fmt.Println("States 1: ", states[int(cmd.Identifier)].Omega1)
			} else if cmd.Trustee == "2" {
				fmt.Println("Omega2 :", omega)
				fmt.Println("States 2: ", states[int(cmd.Identifier)].Omega2)
			}
		}

	default:
		//LOG: fmt.Printf("[TAM Worker %d] Unknown message to %v\n", workerID, cmd)
	}
}
