package trustmodelupdate

import (
	"github.com/vs-uulm/go-subjectivelogic/pkg/subjectivelogic"
	"github.com/vs-uulm/go-taf/pkg/core"
)

type UpdateAtomicTrustOpinion struct {
	opinion     subjectivelogic.QueryableOpinion
	trustSource core.TrustSource
	trustee     string
	trustor     string
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

func (u UpdateAtomicTrustOpinion) Trustor() string {
	return u.trustor
}

func CreateAtomicTrustOpinionUpdate(opinion subjectivelogic.QueryableOpinion, trustor string, trustee string, source core.TrustSource) UpdateAtomicTrustOpinion {
	return UpdateAtomicTrustOpinion{
		opinion:     opinion,
		trustSource: source,
		trustee:     trustee,
		trustor:     trustee,
	}
}

func (u UpdateAtomicTrustOpinion) Type() core.UpdateOp {
	return core.UPDATE_ATO
}
