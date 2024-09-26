package vehiclecomputermigration

import (
	"fmt"
	"github.com/vs-uulm/go-subjectivelogic/pkg/subjectivelogic"
	"github.com/vs-uulm/go-taf/pkg/core"
)

var vc1DTI, _ = subjectivelogic.NewOpinion(0.2, 0.1, 0.7, 0.5)
var vc2DTI, _ = subjectivelogic.NewOpinion(0.2, 0.1, 0.7, 0.5)

var vc1ExistenceWeights = map[core.EvidenceType]float64{
	core.AIV_SECURE_BOOT:                          0.2,
	core.AIV_ACCESS_CONTROL:                       0.2,
	core.AIV_CONTROL_FLOW_INTEGRITY:               0.1,
	core.AIV_SECURE_OTA:                           0.1,
	core.AIV_APPLICATION_ISOLATION:                0.1,
	core.AIV_CONFIGURATION_INTEGRITY_VERIFICATION: 0.1,
}

var vc1OutputWeights = map[core.EvidenceType]int{
	core.AIV_SECURE_BOOT:                          2,
	core.AIV_ACCESS_CONTROL:                       0,
	core.AIV_CONTROL_FLOW_INTEGRITY:               2,
	core.AIV_SECURE_OTA:                           0,
	core.AIV_APPLICATION_ISOLATION:                1,
	core.AIV_CONFIGURATION_INTEGRITY_VERIFICATION: 2,
}

var vc2ExistenceWeights = map[core.EvidenceType]float64{
	core.AIV_SECURE_BOOT:                          0.2,
	core.AIV_ACCESS_CONTROL:                       0.2,
	core.AIV_CONTROL_FLOW_INTEGRITY:               0.1,
	core.AIV_SECURE_OTA:                           0.1,
	core.AIV_APPLICATION_ISOLATION:                0.1,
	core.AIV_CONFIGURATION_INTEGRITY_VERIFICATION: 0.1,
}

var vc2OutputWeights = map[core.EvidenceType]int{
	core.AIV_SECURE_BOOT:                          2,
	core.AIV_ACCESS_CONTROL:                       0,
	core.AIV_CONTROL_FLOW_INTEGRITY:               2,
	core.AIV_SECURE_OTA:                           0,
	core.AIV_APPLICATION_ISOLATION:                1,
	core.AIV_CONFIGURATION_INTEGRITY_VERIFICATION: 2,
}

func quantifier(values map[core.EvidenceType]int, designTimeTrustOp subjectivelogic.QueryableOpinion, existenceWeights map[core.EvidenceType]float64, outputWeights map[core.EvidenceType]int) subjectivelogic.QueryableOpinion {
	sl, _ := subjectivelogic.NewOpinion(.0, .0, 1.0, 0.5)

	fmt.Printf("%+v ", existenceWeights)

	belief := designTimeTrustOp.Belief()
	disbelief := designTimeTrustOp.Disbelief()
	uncertainty := designTimeTrustOp.Uncertainty()

	for control, appraisal := range values {
		delta := existenceWeights[control] * designTimeTrustOp.Uncertainty()

		if appraisal == -1 { // control not implemented
			disbelief = disbelief + delta
			uncertainty = uncertainty - delta
		} else if appraisal == 0 {
			if outputWeights[control] == 0 { // still add belief
				belief = belief + delta
				uncertainty = uncertainty - delta
			} else if outputWeights[control] == 1 { // add disbelief
				disbelief = disbelief + delta
				uncertainty = uncertainty - delta
			} else if outputWeights[control] == 2 { // complete disbelief
				belief = 0.0
				disbelief = 1.0
				uncertainty = 0.0
				break // complete disbelief because negative evidence of critical securityControl
			} else {
				// Invalid weight
				// TODO: Error handling
			}
		} else if appraisal == 1 {
			belief = belief + delta
			uncertainty = uncertainty - delta
		} else {
			// No evidence for the control, e.g. appraisal -2 or no evidence received -> Results in higher uncertainty
		}
	}

	sl.Modify(belief, disbelief, uncertainty, sl.BaseRate())

	return &sl
}

var trustSourceQuantifiers = []core.TrustSourceQuantifier{
	{
		Trustor:     "TAF",
		Trustee:     "VC1",
		Scope:       "VC1",
		TrustSource: core.AIV,
		Evidence:    []core.EvidenceType{core.AIV_SECURE_BOOT, core.AIV_SECURE_OTA, core.AIV_ACCESS_CONTROL, core.AIV_APPLICATION_ISOLATION, core.AIV_CONTROL_FLOW_INTEGRITY, core.AIV_CONFIGURATION_INTEGRITY_VERIFICATION},
		Quantifier: func(m map[core.EvidenceType]int) subjectivelogic.QueryableOpinion {
			return quantifier(m, &vc1DTI, vc1ExistenceWeights, vc1OutputWeights)
		},
	},
	{
		Trustor:     "TAF",
		Trustee:     "VC2",
		Scope:       "VC2",
		TrustSource: core.AIV,
		Evidence:    []core.EvidenceType{core.AIV_SECURE_BOOT, core.AIV_SECURE_OTA, core.AIV_ACCESS_CONTROL, core.AIV_APPLICATION_ISOLATION, core.AIV_CONTROL_FLOW_INTEGRITY, core.AIV_CONFIGURATION_INTEGRITY_VERIFICATION},
		Quantifier: func(m map[core.EvidenceType]int) subjectivelogic.QueryableOpinion {
			return quantifier(m, &vc2DTI, vc2ExistenceWeights, vc2OutputWeights)
		},
	},
}
