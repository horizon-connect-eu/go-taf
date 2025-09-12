package trustsourcehandler

import (
	"github.com/horizon-connect-eu/go-taf/internal/flow/completionhandler"
	"github.com/horizon-connect-eu/go-taf/pkg/command"
	"github.com/horizon-connect-eu/go-taf/pkg/core"
	v2xmsg "github.com/horizon-connect-eu/go-taf/pkg/message/v2x"
	"github.com/horizon-connect-eu/go-taf/pkg/trustmodel/session"
	"github.com/horizon-connect-eu/go-taf/pkg/trustmodel/trustmodelupdate"
	"github.com/vs-uulm/go-subjectivelogic/pkg/subjectivelogic"
	"log/slog"
	"strconv"
)

type NtmHandler struct {
	sessionTsqs                map[string][]core.TrustSourceQuantifier
	latestSubscriptionEvidence map[string]map[core.EvidenceType]int
	tam                        TAMAccess
	logger                     *slog.Logger
}

func CreateNtmHandler(tam TAMAccess, logger *slog.Logger) *NtmHandler {
	return &NtmHandler{
		sessionTsqs:                make(map[string][]core.TrustSourceQuantifier),
		latestSubscriptionEvidence: make(map[string]map[core.EvidenceType]int),
		logger:                     logger,
		tam:                        tam,
	}
}

func (n *NtmHandler) Initialize() {
	return
}

func (n *NtmHandler) AddSession(sess session.Session, handler *completionhandler.CompletionHandler) {
	n.sessionTsqs[sess.ID()] = sess.TrustSourceQuantifiers()
}

func (n *NtmHandler) RemoveSession(sess session.Session, handler *completionhandler.CompletionHandler) {
	delete(n.sessionTsqs, sess.ID())
}

func (n *NtmHandler) TrustSourceType() core.TrustSource {
	return core.NTM
}

func (n *NtmHandler) RegisteredSessions() []string {
	sessions := make([]string, len(n.sessionTsqs))
	i := 0
	for k := range n.sessionTsqs {
		sessions[i] = k
		i++
	}
	return sessions
}

func (n *NtmHandler) HandleNotify(cmd command.HandleNotify[v2xmsg.V2XNtm]) {

	receivedNtmOpinions := make(map[int64]subjectivelogic.QueryableOpinion) //flag trustees with changes

	if nil == cmd.Notify.V2XSourceSet {
		n.logger.Info("Empty NTM Source Set, ignoring V2X_NTM message")
		return
	}

	for _, entry := range cmd.Notify.V2XSourceSet {
		sourceID := entry.V2XSourceID
		belief := entry.Opinion.Belief
		disbelief := entry.Opinion.Disbelief
		uncertainty := entry.Opinion.Uncertainty
		baseRate := entry.Opinion.BaseRate

		opinion, err := subjectivelogic.NewOpinion(belief, disbelief, uncertainty, baseRate)
		if err != nil {
			n.logger.Warn("Invalid subjective logic recevied in V2X_NTM, ignoring opinion", "Belief", belief, "Disbelief", disbelief, "Uncertainty", uncertainty, "BaseRate", baseRate)
			continue
		}

		receivedNtmOpinions[sourceID] = &opinion
	}

	//Iterate over all sessions register for TCH, call quantifiers and relay updates
	for sessionId, tsqs := range n.sessionTsqs {
		sess := n.tam.Sessions()[sessionId]
		//loop through all updated trustees and find fitting quantifiers; if successful, apply quantifier and add ATO update
		updates := make([]core.Update, 0)
		for opinionTarget := range receivedNtmOpinions {
			for _, tsq := range tsqs {
				if tsq.TrustSource != core.NTM {
					continue
				} else if tsq.Trustor == "MEC" && tsq.Trustee == "V_*" {
					//This is a bit bogus as we already have the opinion. But the quantifier might modify it in the future.
					ato := tsq.Quantifier(map[core.EvidenceType]interface{}{
						core.NTM_REMOTE_OPINION: receivedNtmOpinions[opinionTarget],
					})
					n.logger.Debug("Opinion for "+strconv.FormatInt(opinionTarget, 10), "SL", ato.String())
					updates = append(updates, trustmodelupdate.CreateAtomicTrustOpinionUpdate(ato, "V_ego", "V_"+strconv.FormatInt(opinionTarget, 10), core.NTM))
				}
			}
		}
		if len(updates) > 0 {
			for tmiID, fullTmiID := range sess.TrustModelInstances() {
				tmiUpdateCmd := command.CreateHandleTMIUpdate(fullTmiID, cmd.Notify.Tag, updates...)
				n.tam.DispatchToWorker(sess, tmiID, tmiUpdateCmd)
			}
		}
	}
}
