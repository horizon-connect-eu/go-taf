package trustmodelupdate

import (
	"github.com/vs-uulm/go-subjectivelogic/pkg/subjectivelogic"
	"github.com/vs-uulm/go-taf/pkg/core"
)

type UpdateAtomicTrustOpinion struct {
	Opinion                       subjectivelogic.QueryableOpinion
	TrustSourceQuantifierInstance core.TrustSourceQuantifierInstance
}

func (u UpdateAtomicTrustOpinion) Type() core.UpdateOp {
	return core.UPDATE_ATO
}
