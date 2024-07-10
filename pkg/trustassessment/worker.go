package trustassessment

import (
	actualtlee "connect.informatik.uni-ulm.de/coordination/tlee-implementation/pkg/core"
	"fmt"
	"github.com/vs-uulm/go-taf/internal/logger"
	"github.com/vs-uulm/go-taf/internal/util"
	"github.com/vs-uulm/go-taf/pkg/command"
	"github.com/vs-uulm/go-taf/pkg/core"
	tasmsg "github.com/vs-uulm/go-taf/pkg/message/tas"
	internaltlee "github.com/vs-uulm/go-taf/pkg/tlee"
	"github.com/vs-uulm/taf-tlee-interface/pkg/tleeinterface"
	"log/slog"
	"time"
)

type Worker struct {
	tafContext core.RuntimeContext
	id         int
	inputs     <-chan core.Command
	logger     *slog.Logger
	//	states     State
}

func (t *trustAssessmentManager) SpawnNewWorker(id int, inputs <-chan core.Command, tafContext core.RuntimeContext) Worker {
	return Worker{
		tafContext: tafContext,
		id:         id,
		inputs:     inputs,
		//		states:     t.mkStateDatabase(),
		logger: logger.CreateChildLogger(tafContext.Logger, fmt.Sprintf("TAM-WORKER-%d", id)),
	}
}

func (w *Worker) Run() {
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
	//TODO use ctx to shutdown worker
}

//// Processes the messages received via the specified channel as fast as possible.
//func (t *trustAssessmentManager) tamWorker(id int, inputs <-chan Command, tablelogger consolelogger.Logger) {
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
//			processCommand(id, tablelogger, command, states)
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

func (w *Worker) processCommand(cmd core.Command) {

	var doRunTlee = false

	switch cmd := cmd.(type) {
	case command.HandleRequest[tasmsg.TasInitRequest]:
		w.logger.Debug("Got HandleTasInitRequest")
		/*
			case command.InitTMICommand:
				w.logger.Debug("handling InitTMICommand", "Message", fmt.Sprintf("%+v", cmd))

				tmiID = int(cmd.Identifier)
				w.states[tmiID] = trustmodelinstance.NewTrustModelInstance(tmiID, cmd.TrustModelTemplate)

				w.logger.Debug("Trust Model with ID 1139 has been instantiated ")

			case command.UpdateTOCommand:
				w.logger.Debug("handling UpdateATOCommand", "Message", fmt.Sprintf("%+v", cmd))

				trustModelInstance := w.states[int(cmd.Identifier)]

				w.logger.Info("Updating TMI", "TMI ID", trustModelInstance.GetId())

				var evidenceStr string
				if cmd.Evidence {
					evidenceStr = "positive"
				} else {
					evidenceStr = "negative"
				}

				w.logger.LogAttrs(w.tafContext.Context, slog.LevelInfo, "New evidence received", slog.Group("Evidence"),
					slog.String("Trust Source", cmd.TS_ID),
					slog.String("Trust Object", "ECU"+cmd.Trustee),
					slog.String("Evidence", evidenceStr),
				)

				tmiID = int(cmd.Identifier)

				var evidenceCollection map[string]bool
				var omegaDTI subjectivelogic.Opinion
				var omega subjectivelogic.Opinion

				if cmd.Trustee == "1" {
					evidenceCollection = w.states[tmiID].Evidence1
					omegaDTI = w.states[tmiID].Omega_DTI_1
				} else if cmd.Trustee == "2" {
					evidenceCollection = w.states[tmiID].Evidence2
					omegaDTI = w.states[tmiID].Omega_DTI_2
				} else {
					return
				}

				evidenceCollection[cmd.TS_ID] = cmd.Evidence
				omega = omegaDTI

				for tsId, evidence := range evidenceCollection {
					// Equation: delta = u_DTI * weight_ts -> delta specifies how much belief, disbelief and uncertainty will be increased / decreased
					delta := math.Abs(math.Round(omegaDTI.Uncertainty()*w.states[tmiID].Weights[tsId]*100) / 100) // Round delta value to two decimal places to prevent rounding errors in the belief, disbelief and uncertainty values

					if evidence { // positive evidence, e.g. secure boot ran successfully
						omega, _ = subjectivelogic.NewOpinion(omega.Belief()+delta, omega.Disbelief(), omega.Uncertainty()-delta, omega.BaseRate())
					} else if !evidence { // negative evidence, e.g. secure boot didn't run successfully
						omega, _ = subjectivelogic.NewOpinion(omega.Belief(), omega.Disbelief()+delta, omega.Uncertainty()-delta, omega.BaseRate())
					}
				}

				if entry, ok := w.states[int(cmd.Identifier)]; ok {
					// round values to two decimal places
					err := omega.Modify(math.Abs(math.Round(omega.Belief()*100)/100), math.Abs(math.Round(omega.Disbelief()*100)/100), math.Abs(math.Round(omega.Uncertainty()*100)/100), omega.BaseRate())
					if err != nil {
						w.logger.Warn("Failed SL Opinion operation", "Error", err)
					}

					if cmd.Trustee == "1" {
						entry.Omega1 = omega
					} else if cmd.Trustee == "2" {
						entry.Omega2 = omega
					}

					entry.Version = entry.Version + 1

					w.states[int(cmd.Identifier)] = entry

				}

				doRunTlee = true
		*/
	default:
		w.logger.Warn("Unknown message", "Message", fmt.Sprintf("%+v", cmd))
	}

	if doRunTlee {

		//TLEE execution
		var tlee tleeinterface.TLEE
		if w.tafContext.Configuration.TLEE.UseInternalTLEE {
			tlee = &internaltlee.TLEE{}
		} else {
			tlee = &actualtlee.TLEE{}
		}
		util.UNUSED(tlee)
		//tleeResults := tlee.RunTLEE(strconv.Itoa(tmi.Id), tmi.Version, uint32(tmi.Fingerprint), tmi.GetTrustGraphStructure(), tmi.GetTrustRelationships())
	}
}
