package trustsourcehandler

import (
	"fmt"
	"github.com/horizon-connect-eu/go-taf/internal/flow/completionhandler"
	"github.com/horizon-connect-eu/go-taf/pkg/command"
	"github.com/horizon-connect-eu/go-taf/pkg/core"
	aivmsg "github.com/horizon-connect-eu/go-taf/pkg/message/aiv"
	"github.com/horizon-connect-eu/go-taf/pkg/trustmodel/session"
	"github.com/horizon-connect-eu/go-taf/pkg/trustmodel/trustmodelupdate"
	"log/slog"
)

/*
The AivHandler is a trust source handler for AIV-based evidence. This Handler creates an individual subscription at the
AIV for each session, as each session might require different claims in that subscriptions. There is hence a 1:1 mapping
of sessions and AIV subscriptions at runtime.
*/
type AivHandler struct {
	sessionTsqs                  map[string][]core.TrustSourceQuantifier
	latestSubscriptionEvidence   map[string]map[string]map[core.EvidenceType]interface{}
	tam                          TAMAccess
	tsm                          TSMAccess
	logger                       *slog.Logger
	aivSubscriptionIDtoSession   map[string]session.Session
	sessionIDtoAivSubscriptionID map[string]string
}

func CreateAivHandler(tam TAMAccess, tsm TSMAccess, logger *slog.Logger) *AivHandler {
	return &AivHandler{
		sessionTsqs:                  make(map[string][]core.TrustSourceQuantifier),
		latestSubscriptionEvidence:   make(map[string]map[string]map[core.EvidenceType]interface{}),
		logger:                       logger,
		tsm:                          tsm,
		tam:                          tam,
		aivSubscriptionIDtoSession:   make(map[string]session.Session),
		sessionIDtoAivSubscriptionID: make(map[string]string),
	}
}

func (h *AivHandler) Initialize() {
	return
}

func (h *AivHandler) AddSession(sess session.Session, handler *completionhandler.CompletionHandler) {
	h.tsm.SubscribeAIV(handler, sess)
}

func (h *AivHandler) RegisterSubscription(sess session.Session, subscriptionID string) {
	h.latestSubscriptionEvidence[subscriptionID] = make(map[string]map[core.EvidenceType]interface{})
	h.aivSubscriptionIDtoSession[subscriptionID] = sess
	h.sessionIDtoAivSubscriptionID[sess.ID()] = subscriptionID
}

func (h *AivHandler) RemoveSession(sess session.Session, handler *completionhandler.CompletionHandler) {
	subId, exists := h.sessionIDtoAivSubscriptionID[sess.ID()]
	if !exists {
		h.logger.Warn("Unknown session for AIV_NOTIFY, discarding message", "Session ID", sess.ID())
		return
	} else {
		h.tsm.UnsubscribeAIV(subId, handler)
		delete(h.sessionIDtoAivSubscriptionID, sess.ID())
		delete(h.aivSubscriptionIDtoSession, subId)
		delete(h.latestSubscriptionEvidence, subId)
	}
}

func (h *AivHandler) TrustSourceType() core.TrustSource {
	return core.AIV
}

func (h *AivHandler) RegisteredSessions() []string {
	sessions := make([]string, len(h.sessionTsqs))
	i := 0
	for k := range h.sessionTsqs {
		sessions[i] = k
		i++
	}
	return sessions
}

func (h *AivHandler) HandleNotify(cmd command.HandleNotify[aivmsg.AivNotify]) {
	//Check whether the subscription is known, otherwise return
	subID := cmd.Notify.SubscriptionID
	sess, exists := h.aivSubscriptionIDtoSession[subID]
	if !exists {
		h.logger.Warn("Unknown subscription for AIV_NOTIFY, discarding message", "Subscription ID", cmd.Notify.SubscriptionID)
		return
	}

	//Extract raw evidence from the message and store it into latestEvidence
	updatedTrustees := make(map[string]bool)
	for _, trusteeReport := range cmd.Notify.TrusteeReports {
		trusteeID := *trusteeReport.TrusteeID
		//Discard old evidence and always create a new map
		h.latestSubscriptionEvidence[subID][trusteeID] = make(map[core.EvidenceType]interface{})
		for _, attestationReport := range trusteeReport.AttestationReport {
			evidenceType := core.EvidenceTypeBySourceAndName(core.AIV, attestationReport.Claim)
			value := int(attestationReport.Appraisal)
			h.latestSubscriptionEvidence[subID][trusteeID][evidenceType] = value
			updatedTrustees[trusteeID] = true
		}
	}

	//loop through all updated trustees and find fitting quantifiers; if successful, apply quantifier and add ATO update
	updates := make([]core.Update, 0)
	for trustee := range updatedTrustees {
		for _, tsq := range sess.TrustSourceQuantifiers() {
			if tsq.TrustSource != core.AIV {
				break
			} else if tsq.Trustee == trustee {
				ato := tsq.Quantifier(h.latestSubscriptionEvidence[subID][trustee])
				h.logger.Debug("Opinion for "+trustee, "SL", ato.String(), "Input", fmt.Sprintf("%v", h.latestSubscriptionEvidence[subID][trustee]))
				updates = append(updates, trustmodelupdate.CreateAtomicTrustOpinionUpdate(ato, "", trustee, core.AIV))
			}
		}
	}
	if len(updates) > 0 {
		for tmiID, fullTmiID := range sess.TrustModelInstances() {
			tmiUpdateCmd := command.CreateHandleTMIUpdate(fullTmiID, cmd.Notify.Tag, updates...)
			h.tam.DispatchToWorker(sess, tmiID, tmiUpdateCmd)
		}
	}
}
