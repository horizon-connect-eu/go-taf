package trustassessment

import (
	"context"
	"fmt"
	"github.com/pterm/pterm"
	"github.com/vs-uulm/go-taf/internal/consolelogger"
	"github.com/vs-uulm/go-taf/pkg/trustdecision"
	"github.com/vs-uulm/go-taf/pkg/trustmodel/trustmodelinstance"
	"github.com/vs-uulm/taf-tlee-interface/pkg/tlee"
	"math"
	"strconv"
	"time"

	"github.com/vs-uulm/taf-tlee-interface/pkg/subjectivelogic"
)

type Worker struct {
	id     int
	inputs <-chan Command
	logger consolelogger.Logger
	states State
}

func (t *trustAssessmentManager) SpawnNewWorker(id int, inputs <-chan Command, logger consolelogger.Logger) Worker {
	return Worker{
		id:     id,
		inputs: inputs,
		logger: logger,
		states: t.mkStateDatabase(),
	}
}

func (w *Worker) Run(ctx context.Context) {
	// Ticker for latency benchmark
	latTicker := time.NewTicker(1 * time.Second)
	latMeasurePending := false

	for {
		select {
		case command := <-w.inputs:
			w.processCommand(command)
			if latMeasurePending && w.id == 0 {
				//fmt.Printf("TAM: latency of %d µs\n", time.Since(command.Timestamp).Microseconds())
				latMeasurePending = false
			}
		case <-latTicker.C:
			latMeasurePending = true

		}
	}
}

//// Processes the messages received via the specified channel as fast as possible.
//func (t *trustAssessmentManager) tamWorker(id int, inputs <-chan Command, logger consolelogger.Logger) {
//	states := t.mkStateDatabase()
//	//results := t.mkResultsDatabase()
//
//	// Ticker for latency benchmark
//	latTicker := time.NewTicker(1 * time.Second)
//	latMeasurePending := false
//
//	for {
//		select {
//		case command := <-inputs:
//			processCommand(id, logger, command, states)
//			if latMeasurePending && id == 0 {
//				//fmt.Printf("TAM: latency of %d µs\n", time.Since(command.Timestamp).Microseconds())
//				latMeasurePending = false
//			}
//		case <-latTicker.C:
//			latMeasurePending = true
//
//		}
//	}
//}

func (w *Worker) processCommand(cmd Command) {

	var doRunTlee = false
	var tmiID int

	switch cmd := cmd.(type) {
	case InitTMICommand:
		//LOG: fmt.Printf("[TAM Worker %d] handling InitTMICommand: %v\n", workerID, cmd)

		tmiID = int(cmd.Identifier)
		w.states[tmiID] = trustmodelinstance.NewTrustModelInstance(tmiID, cmd.TrustModelTemplate)

	case UpdateTOCommand:
		//LOG: fmt.Printf("[TAM Worker %d] handling UpdateATOCommand: %v\n", workerID, cmd)

		//trustModelInstance := states[int(cmd.Identifier)]

		//LOG: fmt.Printf("[TAM Worker %d] updating TMI %d\n", workerID, trustModelInstance.GetId())

		//w.logger.Info("New evidence received: (Trust Source: " + cmd.TS_ID + "; Trust Object: ECU" + cmd.Trustee + "; Evidence: " + strconv.FormatBool(cmd.Evidence) + ")")

		var evidenceStr string
		if cmd.Evidence {
			evidenceStr = strconv.FormatBool(cmd.Evidence)
		} else {
			evidenceStr = pterm.Red(strconv.FormatBool(cmd.Evidence))
		}

		w.logger.InfoWithArgs("New evidence received", pterm.LoggerArgument{
			Key:   "Trust Source",
			Value: cmd.TS_ID,
		}, pterm.LoggerArgument{
			Key:   "Trust Object",
			Value: "ECU" + cmd.Trustee,
		}, pterm.LoggerArgument{
			Key:   "Evidence",
			Value: evidenceStr,
		})

		tmiID = int(cmd.Identifier)

		var evidence_collection map[string]bool
		var omega_DTI subjectivelogic.Opinion
		var omega subjectivelogic.Opinion

		if cmd.Trustee == "1" {
			evidence_collection = w.states[tmiID].Evidence1
			omega_DTI = w.states[tmiID].Omega_DTI_1
		} else if cmd.Trustee == "2" {
			evidence_collection = w.states[tmiID].Evidence2
			omega_DTI = w.states[tmiID].Omega_DTI_2
		} else {
			return
		}

		evidence_collection[cmd.TS_ID] = cmd.Evidence
		omega = omega_DTI

		for ts_id, evidence := range evidence_collection {
			// Equation: delta = u_DTI * weight_ts -> delta specifies how much belief, disbelief and uncertainty will be increased / decreased
			if evidence { // positive evidence, e.g. secure boot ran successfully
				omega.Belief = omega.Belief + omega_DTI.Uncertainty*w.states[tmiID].Weights[ts_id]
				omega.Uncertainty = omega.Uncertainty - omega_DTI.Uncertainty*w.states[tmiID].Weights[ts_id]
			} else if !evidence { // negative evidence, e.g. secure boot didn't run successfully
				omega.Disbelief = omega.Disbelief + omega_DTI.Uncertainty*w.states[tmiID].Weights[ts_id]
				omega.Uncertainty = omega.Uncertainty - omega_DTI.Uncertainty*w.states[tmiID].Weights[ts_id]
			}
		}

		if entry, ok := w.states[int(cmd.Identifier)]; ok {
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

			w.states[int(cmd.Identifier)] = entry

		}

		doRunTlee = true
		//logger.Info(pterm.Blue("New evidence received for TMI 1139"))
		//logger.Warn(pterm.Blue("New evidence received for TMI 1139"))

	default:
		//LOG: fmt.Printf("[TAM Worker %d] Unknown message to %v\n", workerID, cmd)
	}

	if doRunTlee {

		var tmi = w.states[tmiID]

		var tleeResults = tlee.RunTLEE(strconv.Itoa(tmi.Id), tmi.Version, uint32(tmi.Fingerprint), tmi.GetStructure(), tmi.GetValues())

		//map[string]subjectivelogic.Opinion

		//TDE
		var tdeResults = make(map[string]bool)

		tdeResults["1139-123"] = trustdecision.Decide(tleeResults["1139-123"], tmi.RTL1)
		tdeResults["1139-124"] = trustdecision.Decide(tleeResults["1139-124"], tmi.RTL2)

		projectedRtls := map[string]float64{
			"1139-123": trustdecision.ProjectProbability(tmi.RTL1),
			"1139-124": trustdecision.ProjectProbability(tmi.RTL2),
		}

		rtls := map[string]subjectivelogic.Opinion{
			"1139-123": tmi.RTL1,
			"1139-124": tmi.RTL2,
		}

		trustee := map[string]string{
			"1139-123": "ECU1",
			"1139-124": "ECU2",
		}

		//print table only after all evidences are set for both trust objects (2*3)
		if len(tmi.Evidence1)+len(tmi.Evidence2) >= 6 {
			w.logger.Info("Result of TLEE and TDE Execution:")
			printTable(w.logger, tleeResults, tdeResults)

			for _, id := range []string{"1139-123", "1139-124"} {
				if !tdeResults[id] {
					w.logger.WarnWithArgs(pterm.Red(trustee[id]+" is untrustworthy!"), pterm.LoggerArgument{
						Key:   "ATL",
						Value: tleeResults[id].ToString() + " ==> Projected Probability: " + pterm.Red(fmt.Sprintf("%.2f", trustdecision.ProjectProbability(tleeResults[id]))),
					}, pterm.LoggerArgument{
						Key:   "RTL",
						Value: rtls[id].ToString() + " ==> Projected Probability: " + fmt.Sprintf("%.2f", projectedRtls[id]),
					})
				} else {
					w.logger.InfoWithArgs(trustee[id]+" is trustworthy", pterm.LoggerArgument{
						Key:   "ATL",
						Value: tleeResults[id].ToString() + " ==> Projected Probability: " + pterm.Green(fmt.Sprintf("%.2f", trustdecision.ProjectProbability(tleeResults[id]))),
					}, pterm.LoggerArgument{
						Key:   "RTL",
						Value: rtls[id].ToString() + " ==> Projected Probability: " + fmt.Sprintf("%.2f", projectedRtls[id]),
					})

				}
			}
		}
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
