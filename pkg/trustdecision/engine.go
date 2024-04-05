package trustdecision

import (
	"github.com/vs-uulm/taf-tlee-interface/pkg/subjectivelogic"
)

func Decide(atl subjectivelogic.Opinion, rtl subjectivelogic.Opinion) bool {

	var probabilisticAtl = atl.Belief + atl.Uncertainty*atl.BaseRate
	var probabilisticRtl = rtl.Belief + rtl.Uncertainty*rtl.BaseRate

	return probabilisticAtl > probabilisticRtl
}
