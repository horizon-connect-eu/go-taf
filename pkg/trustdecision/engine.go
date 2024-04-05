package trustdecision

import (
	"github.com/vs-uulm/taf-tlee-interface/pkg/subjectivelogic"
)

func Decide(atl subjectivelogic.Opinion, rtl subjectivelogic.Opinion) bool {

	var probabilisticAtl = ProjectProbability(atl)
	var probabilisticRtl = ProjectProbability(rtl)

	return probabilisticAtl > probabilisticRtl
}

func ProjectProbability(opinion subjectivelogic.Opinion) float64 {
	return opinion.Belief + opinion.Uncertainty*opinion.BaseRate
}
