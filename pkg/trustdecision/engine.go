package trustdecision

import (
	"github.com/vs-uulm/go-subjectivelogic/pkg/subjectivelogic"
	"github.com/vs-uulm/go-taf/pkg/core"
)

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
