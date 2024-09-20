package trustmodelupdate

import (
	"github.com/vs-uulm/go-subjectivelogic/pkg/subjectivelogic"
	"github.com/vs-uulm/go-taf/pkg/core"
)

type UpdateAtomicTrustOpinion struct {
	opinion     subjectivelogic.QueryableOpinion
	trustSource core.TrustSource
	trustee     string
}

func (u UpdateAtomicTrustOpinion) Opinion() subjectivelogic.QueryableOpinion {
	return u.opinion
}

func (u UpdateAtomicTrustOpinion) TrustSource() core.TrustSource {
	return u.trustSource
}

func (u UpdateAtomicTrustOpinion) Trustee() string {
	return u.trustee
}

func CreateAtomicTrustOpinionUpdate(opinion subjectivelogic.QueryableOpinion, trustee string, source core.TrustSource) UpdateAtomicTrustOpinion {
	return UpdateAtomicTrustOpinion{
		opinion:     opinion,
		trustSource: source,
		trustee:     trustee,
	}
}

func (u UpdateAtomicTrustOpinion) Type() core.UpdateOp {
	return core.UPDATE_ATO
}
