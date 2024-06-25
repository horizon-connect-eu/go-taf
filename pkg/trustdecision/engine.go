package trustdecision

import (
	"github.com/vs-uulm/go-subjectivelogic/pkg/subjectivelogic"
)

func Decide(atl subjectivelogic.QueryableOpinion, rtl subjectivelogic.QueryableOpinion) bool {

	var probabilisticAtl = ProjectProbability(atl)
	var probabilisticRtl = ProjectProbability(rtl)

	return probabilisticAtl > probabilisticRtl
}

func ProjectProbability(opinion subjectivelogic.QueryableOpinion) float64 {
	return opinion.Belief() + opinion.Uncertainty()*opinion.BaseRate()
}
