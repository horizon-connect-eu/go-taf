package trustassessment

import (
	"github.com/pterm/pterm"
	"github.com/vs-uulm/go-taf/internal/consolelogger"
	"github.com/vs-uulm/go-taf/pkg/trustdecision"
	"github.com/vs-uulm/taf-tlee-interface/pkg/tlee"
	"math"
	"strconv"
	"time"

	"github.com/vs-uulm/go-taf/pkg/trustmodel/instance"
	"github.com/vs-uulm/taf-tlee-interface/pkg/subjectivelogic"
)

// Processes the messages received via the specified channel as fast as possible.
func (t *trustAssessmentManager) tamWorker(id int, inputs <-chan Command, logger consolelogger.Logger) {
	states := t.mkStateDatabase()
	//results := t.mkResultsDatabase()

	// Ticker for latency benchmark
	latTicker := time.NewTicker(1 * time.Second)
	latMeasurePending := false

	for {
		select {
		case command := <-inputs:
			processCommand(id, logger, command, states)
			if latMeasurePending && id == 0 {
				//fmt.Printf("TAM: latency of %d µs\n", time.Since(command.Timestamp).Microseconds())
				latMeasurePending = false
			}
		case <-latTicker.C:
			latMeasurePending = true

		}
	}
}

func processCommand(workerID int, logger consolelogger.Logger, cmd Command, states State) {

	var doRunTlee = false
	var tmiID int

	switch cmd := cmd.(type) {
	case InitTMICommand:
		//LOG: fmt.Printf("[TAM Worker %d] handling InitTMICommand: %v\n", workerID, cmd)

		tmiID = int(cmd.Identifier)
		states[tmiID] = instance.NewTrustModelInstance(tmiID, cmd.TrustModelTemplate)

	case UpdateTOCommand:
		//LOG: fmt.Printf("[TAM Worker %d] handling UpdateATOCommand: %v\n", workerID, cmd)

		//trustModelInstance := states[int(cmd.Identifier)]

		//LOG: fmt.Printf("[TAM Worker %d] updating TMI %d\n", workerID, trustModelInstance.GetId())

		logger.Info("New evidence received: (Trust Source: " + cmd.TS_ID + "; Trust Object: ECU" + cmd.Trustee + "; Evidence: " + strconv.FormatBool(cmd.Evidence) + ")")

		tmiID = int(cmd.Identifier)

		var evidence_collection map[string]bool
		var omega_DTI subjectivelogic.Opinion
		var omega subjectivelogic.Opinion

		if cmd.Trustee == "1" {
			evidence_collection = states[tmiID].Evidence1
			omega_DTI = states[tmiID].Omega_DTI_1
		} else if cmd.Trustee == "2" {
			evidence_collection = states[tmiID].Evidence2
			omega_DTI = states[tmiID].Omega_DTI_2
		} else {
			return
		}

		evidence_collection[cmd.TS_ID] = cmd.Evidence
		omega = omega_DTI

		for ts_id, evidence := range evidence_collection {
			// Equation: delta = u_DTI * weight_ts -> delta specifies how much belief, disbelief and uncertainty will be increased / decreased
			if evidence { // positive evidence, e.g. secure boot ran successfully
				//TODO for Artur: replace with `tmiID`
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

			entry.Version = entry.Version + 1

			states[int(cmd.Identifier)] = entry

		}

		doRunTlee = true
		//logger.Info(pterm.Blue("New evidence received for TMI 1139"))
		//logger.Warn(pterm.Blue("New evidence received for TMI 1139"))

	default:
		//LOG: fmt.Printf("[TAM Worker %d] Unknown message to %v\n", workerID, cmd)
	}

	if doRunTlee {

		var tmi = states[tmiID]

		var tleeResults = tlee.RunTLEE(strconv.Itoa(tmi.Id), tmi.Version, uint32(tmi.Fingerprint), tmi.GetStructure(), tmi.GetValues())

		//map[string]subjectivelogic.Opinion

		//TDE
		var tdeResults = make(map[string]bool)

		tdeResults["1139-123"] = trustdecision.Decide(tleeResults["1139-123"], tmi.RTL1)
		tdeResults["1139-124"] = trustdecision.Decide(tleeResults["1139-124"], tmi.RTL2)

		//print Table?
		logger.Info("Executed TLEE and TDE:")
		printTable(logger, tleeResults, tdeResults)
	}
}

func printTable(logger consolelogger.Logger, atls map[string]subjectivelogic.Opinion, tds map[string]bool) {

	logger.Table([][]string{
		{"Rel. ID", "Trustor", "Trustee", "ATL", "Trust Decision"},
		{"1139-123", "TAF", "ECU1", atls["1139-123"].ToString(), printTDE(tds["1139-123"])},
		{"1139-124", "TAF", "ECU2", atls["1139-124"].ToString(), printTDE(tds["1139-124"])},
	})
}

func printTDE(value bool) string {
	if value {
		return pterm.Green(" ✔ ")
	} else {
		return pterm.Red(" ✗ ")
	}
}
