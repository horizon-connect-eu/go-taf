package trustdecision

import (
	"github.com/vs-uulm/go-subjectivelogic/pkg/subjectivelogic"
	"github.com/vs-uulm/go-taf/pkg/core"
)

/*
Decide produces the final core.TrustDecision based on the actual and requested trust levels. There are three potential
results: core.TRUSTWORTHY, core.NOT_TRUSTWORTHY, and core.UNDECIDABLE.
The latter one is used in case uncertainty is too high to decide upon the trustworthiness.
*/
func Decide(atl subjectivelogic.QueryableOpinion, rtl subjectivelogic.QueryableOpinion) core.TrustDecision {
	if atl.Uncertainty() == 1 {
		return core.UNDECIDABLE
	} else {
		var probabilisticAtl = ProjectProbability(atl)
		var probabilisticRtl = ProjectProbability(rtl)
		if probabilisticAtl > probabilisticRtl {
			return core.TRUSTWORTHY
		} else {
			return core.NOT_TRUSTWORTHY
		}
	}
}

func ProjectProbability(opinion subjectivelogic.QueryableOpinion) float64 {
	return opinion.Belief() + opinion.Uncertainty()*opinion.BaseRate()
}
