package trustmodelupdate

import (
	"github.com/vs-uulm/go-subjectivelogic/pkg/subjectivelogic"
	"github.com/vs-uulm/go-taf/pkg/core"
)

// TODO make member variables private
type UpdateAtomicTrustOpinion struct {
	Opinion     subjectivelogic.QueryableOpinion
	trustSource core.TrustSource
	Trustee     string
}

func CreateAtomicTrustOpinionUpdate(opinion subjectivelogic.QueryableOpinion, trustee string, source core.TrustSource) UpdateAtomicTrustOpinion {
	return UpdateAtomicTrustOpinion{
		Opinion:     opinion,
		trustSource: source,
		Trustee:     trustee,
	}
}

func (u UpdateAtomicTrustOpinion) Type() core.UpdateOp {
	return core.UPDATE_ATO
}
