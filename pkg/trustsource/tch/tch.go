package tch

import (
	"fmt"
	"github.com/vs-uulm/go-taf/internal/flow/completionhandler"
	"github.com/vs-uulm/go-taf/pkg/command"
	"github.com/vs-uulm/go-taf/pkg/core"
	"github.com/vs-uulm/go-taf/pkg/manager"
	tchmsg "github.com/vs-uulm/go-taf/pkg/message/tch"
	"github.com/vs-uulm/go-taf/pkg/trustmodel/session"
	"github.com/vs-uulm/go-taf/pkg/trustmodel/trustmodelupdate"
	"log/slog"
)

type TchHandler struct {
	sessionTsqs                map[string][]core.TrustSourceQuantifier
	latestSubscriptionEvidence map[string]map[core.EvidenceType]int
	tam                        manager.TrustAssessmentManager
	tmm                        manager.TrustModelManager
	tsm                        manager.TrustSourceManager
	logger                     *slog.Logger
}

func CreateTchHandler(tam manager.TrustAssessmentManager, tmm manager.TrustModelManager, tsm manager.TrustSourceManager, logger *slog.Logger) *TchHandler {
	return &TchHandler{
		sessionTsqs:                make(map[string][]core.TrustSourceQuantifier),
		latestSubscriptionEvidence: make(map[string]map[core.EvidenceType]int),
		logger:                     logger,
		tam:                        tam,
		tmm:                        tmm,
		tsm:                        tsm,
	}
}

func (h *TchHandler) Initialize() {
	return
}

func (h *TchHandler) AddSession(sess session.Session, handler *completionhandler.CompletionHandler) {
	h.sessionTsqs[sess.ID()] = sess.TrustSourceQuantifiers()
}

func (h *TchHandler) RemoveSession(sess session.Session, handler *completionhandler.CompletionHandler) {
	delete(h.sessionTsqs, sess.ID())
}

func (h *TchHandler) TrustSourceType() core.TrustSource {
	return core.TCH
}

func (h *TchHandler) RegisteredSessions() []string {
	sessions := make([]string, len(h.sessionTsqs))
	i := 0
	for k := range h.sessionTsqs {
		sessions[i] = k
		i++
	}
	return sessions
}

func (h *TchHandler) HandleNotify(cmd command.HandleNotify[tchmsg.TchNotify]) {
	//Extract raw evidence from the message and store it into latestEvidence
	trusteeID := cmd.Notify.TchReport.TrusteeID
	updatedTrustees := make(map[string]bool)
	for _, trusteeReport := range cmd.Notify.TchReport.TrusteeReports {
		componentID := trusteeReport.ComponentID
		id := trusteeID
		if componentID != nil {
			id = fmt.Sprintf("%s_%s", trusteeID, *componentID)
		}
		//Discard old evidence and always create a new map
		h.latestSubscriptionEvidence[id] = make(map[core.EvidenceType]int)
		for _, attestationReport := range trusteeReport.AttestationReport {
			evidenceType := core.EvidenceTypeByName(attestationReport.Claim)
			value := int(attestationReport.Appraisal)
			h.latestSubscriptionEvidence[id][evidenceType] = value
			updatedTrustees[id] = true
		}
	}

	//Iterate over all sessions register for TCH, call quantifiers and relay updates
	for sessionId, tsqs := range h.sessionTsqs {
		sess := h.tam.Sessions()[sessionId]
		//loop through all updated trustees and find fitting quantifiers; if successful, apply quantifier and add ATO update
		updates := make([]core.Update, 0)
		for trustee := range updatedTrustees {
			for _, tsq := range tsqs {
				if tsq.Trustor == "V_ego" && tsq.Trustee == "V_*" {
					ato := tsq.Quantifier(h.latestSubscriptionEvidence[trustee])
					h.logger.Debug("Opinion for "+trustee, "SL", ato.String(), "Input", fmt.Sprintf("%v", h.latestSubscriptionEvidence[trustee]))
					updates = append(updates, trustmodelupdate.CreateAtomicTrustOpinionUpdate(ato, "V_ego", "V_"+trustee, core.TCH))
				}
			}
		}
		if len(updates) > 0 {
			for tmiID, fullTmiID := range sess.TrustModelInstances() {
				tmiUpdateCmd := command.CreateHandleTMIUpdate(fullTmiID, updates...)
				h.tam.DispatchToWorker(sess, tmiID, tmiUpdateCmd)
			}
		}
	}
}
