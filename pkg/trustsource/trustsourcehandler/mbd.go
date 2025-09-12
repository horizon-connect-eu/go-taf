package trustsourcehandler

import (
	"fmt"
	"github.com/horizon-connect-eu/go-taf/internal/flow/completionhandler"
	"github.com/horizon-connect-eu/go-taf/pkg/command"
	"github.com/horizon-connect-eu/go-taf/pkg/core"
	mbdmsg "github.com/horizon-connect-eu/go-taf/pkg/message/mbd"
	"github.com/horizon-connect-eu/go-taf/pkg/trustmodel/session"
	"github.com/horizon-connect-eu/go-taf/pkg/trustmodel/trustmodelupdate"
	"log/slog"
)

/*
The MbdHandler is a trust source handler for MBD-based evidence. This Handler creates a single subscription at the
misbehavior detection system, if there is at least a single session that requires the MBD as trust source.
There is hence a 1:N mapping the MBD subscription and sessions at runtime.
*/

type MbdHandler struct {
	sessionTsqs                map[string][]core.TrustSourceQuantifier
	latestSubscriptionEvidence map[string]map[core.EvidenceType]interface{}
	tam                        TAMAccess
	tsm                        TSMAccess
	logger                     *slog.Logger
	mbdSubscriptionID          string
	subscriptionState          SubscriptionState
}

func CreateMbdHandler(tam TAMAccess, tsm TSMAccess, logger *slog.Logger) *MbdHandler {
	return &MbdHandler{
		sessionTsqs:                make(map[string][]core.TrustSourceQuantifier),
		latestSubscriptionEvidence: make(map[string]map[core.EvidenceType]interface{}),
		logger:                     logger,
		tsm:                        tsm,
		tam:                        tam,
		subscriptionState:          NA,
	}
}

func (h *MbdHandler) Initialize() {
	return
}

func (h *MbdHandler) AddSession(sess session.Session, handler *completionhandler.CompletionHandler) {
	h.sessionTsqs[sess.ID()] = sess.TrustSourceQuantifiers()
	//TODO: better handling of single subscription: check state
	if len(h.sessionTsqs) == 1 {
		h.tsm.SubscribeMBD(handler)
	}
}

func (h *MbdHandler) RemoveSession(sess session.Session, handler *completionhandler.CompletionHandler) {
	delete(h.sessionTsqs, sess.ID())
	//TODO: better handling of single subscription: check state
	if len(h.sessionTsqs) == 0 {
		h.tsm.UnsubscribeMBD(h.mbdSubscriptionID, handler)
	}
}

func (h *MbdHandler) SetSubscriptionId(id string) {
	h.mbdSubscriptionID = id
	h.subscriptionState = SUBSCRIBED

}

func (h *MbdHandler) SetSubscriptionState(state SubscriptionState) {
	h.subscriptionState = state

}

func (h *MbdHandler) TrustSourceType() core.TrustSource {
	return core.MBD
}

func (h *MbdHandler) RegisteredSessions() []string {
	sessions := make([]string, len(h.sessionTsqs))
	i := 0
	for k := range h.sessionTsqs {
		sessions[i] = k
		i++
	}
	return sessions
}

func (h *MbdHandler) HandleNotify(cmd command.HandleNotify[mbdmsg.MBDNotify]) {
	//Check whether subscription is known, otherwise return
	if cmd.Notify.SubscriptionID != h.mbdSubscriptionID {
		h.logger.Warn("Unknown subscription for MBD_NOTIFY, discarding message", "Subscription ID", cmd.Notify.SubscriptionID)
		return
	}

	updatedTrustees := make(map[string]bool)
	sourceID := cmd.Notify.CpmReport.Content.V2XPduEvidence.SourceID
	for _, observation := range cmd.Notify.CpmReport.Content.ObservationSet {
		id := fmt.Sprintf("C_%d_%d", int(sourceID), int(observation.TargetID))
		//Discard old evidence and always create a new map

		h.latestSubscriptionEvidence[id] = make(map[core.EvidenceType]interface{})
		h.latestSubscriptionEvidence[id][core.MBD_MISBEHAVIOR_REPORT] = int(observation.Check)
		updatedTrustees[id] = true
	}

	//Iterate over all sessions register for MBD, call quantifiers and relay updates
	for sessionId, tsqs := range h.sessionTsqs {
		sess := h.tam.Sessions()[sessionId]
		//loop through all updated trustees and find fitting quantifiers; if successful, apply quantifier and add ATO update
		updates := make([]core.Update, 0)
		for trustee := range updatedTrustees {
			for _, tsq := range tsqs {
				if tsq.TrustSource != core.MBD {
					continue
				} else if tsq.Trustor == "V_ego" && tsq.Trustee == "C_*_*" {
					ato := tsq.Quantifier(h.latestSubscriptionEvidence[trustee])
					h.logger.Debug("Opinion for "+trustee, "SL", ato.String(), "Input", fmt.Sprintf("%v", h.latestSubscriptionEvidence[trustee]))
					updates = append(updates, trustmodelupdate.CreateAtomicTrustOpinionUpdate(ato, tsq.Trustor, trustee, core.MBD))
				} else if tsq.Trustor == "MEC" && tsq.Trustee == "vehicle_*" {
					ato := tsq.Quantifier(h.latestSubscriptionEvidence[trustee])
					h.logger.Debug("Opinion for "+trustee, "SL", ato.String(), "Input", fmt.Sprintf("%v", h.latestSubscriptionEvidence[trustee]))
					updates = append(updates, trustmodelupdate.CreateAtomicTrustOpinionUpdate(ato, tsq.Trustor, fmt.Sprintf("vehicle_%d", int(sourceID)), core.MBD))
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
}
