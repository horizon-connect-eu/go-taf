package trustsourcehandler

import (
	"github.com/vs-uulm/go-subjectivelogic/pkg/subjectivelogic"
	"github.com/vs-uulm/go-taf/internal/flow/completionhandler"
	"github.com/vs-uulm/go-taf/internal/util"
	"github.com/vs-uulm/go-taf/pkg/command"
	"github.com/vs-uulm/go-taf/pkg/core"
	v2xmsg "github.com/vs-uulm/go-taf/pkg/message/v2x"
	"github.com/vs-uulm/go-taf/pkg/trustmodel/session"
	"log/slog"
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

		n.logger.Warn("Created opinion", "sl", opinion)

		util.UNUSED(sourceID, opinion)

	}
	//

	//TODO implement me
	//	panic("implement me")
}
