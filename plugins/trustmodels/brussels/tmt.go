package brussels

import (
	"github.com/vs-uulm/go-subjectivelogic/pkg/subjectivelogic"
	"github.com/vs-uulm/go-taf/pkg/core"
)

func quantifier(values map[core.EvidenceType]int, designTimeTrustOp subjectivelogic.QueryableOpinion, existenceWeights map[core.EvidenceType]float64, outputWeights map[core.EvidenceType]int) subjectivelogic.QueryableOpinion {
	sl, _ := subjectivelogic.NewOpinion(.0, .0, 1.0, 0.5)

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
	core.TrustSourceQuantifier{
		Trustor:     "TAF",
		Trustee:     "VC1",
		Scope:       "VC1",
		TrustSource: core.AIV,
		Evidence:    []core.EvidenceType{core.AIV_SECURE_BOOT, core.AIV_SECURE_OTA, core.AIV_ACCESS_CONTROL, core.AIV_APPLICATION_ISOLATION, core.AIV_CONTROL_FLOW_INTEGRITY},
		Quantifier: func(m map[core.EvidenceType]int) subjectivelogic.QueryableOpinion {
			existenceWeights := map[core.EvidenceType]float64{
				core.AIV_SECURE_BOOT:                          0.2,
				core.AIV_ACCESS_CONTROL:                       0.2,
				core.AIV_CONTROL_FLOW_INTEGRITY:               0.1,
				core.AIV_SECURE_OTA:                           0.1,
				core.AIV_APPLICATION_ISOLATION:                0.1,
				core.AIV_CONFIGURATION_INTEGRITY_VERIFICATION: 0.1,
			}
			outputWeights := map[core.EvidenceType]int{
				core.AIV_SECURE_BOOT:                          2,
				core.AIV_ACCESS_CONTROL:                       0,
				core.AIV_CONTROL_FLOW_INTEGRITY:               2,
				core.AIV_SECURE_OTA:                           0,
				core.AIV_APPLICATION_ISOLATION:                1,
				core.AIV_CONFIGURATION_INTEGRITY_VERIFICATION: 2,
			}
			designTimeTrustOpinion, _ := subjectivelogic.NewOpinion(0.2, 0.1, 0.7, 0.5)
			return quantifier(m, &designTimeTrustOpinion, existenceWeights, outputWeights)
		},
	},
	core.TrustSourceQuantifier{
		Trustor:     "TAF",
		Trustee:     "VC2",
		Scope:       "VC2",
		TrustSource: core.AIV,
		Evidence:    []core.EvidenceType{core.AIV_SECURE_BOOT, core.AIV_SECURE_OTA, core.AIV_ACCESS_CONTROL, core.AIV_APPLICATION_ISOLATION, core.AIV_CONTROL_FLOW_INTEGRITY},
		Quantifier: func(m map[core.EvidenceType]int) subjectivelogic.QueryableOpinion {
			existenceWeights := map[core.EvidenceType]float64{
				core.AIV_SECURE_BOOT:                          0.2,
				core.AIV_ACCESS_CONTROL:                       0.2,
				core.AIV_CONTROL_FLOW_INTEGRITY:               0.1,
				core.AIV_SECURE_OTA:                           0.1,
				core.AIV_APPLICATION_ISOLATION:                0.1,
				core.AIV_CONFIGURATION_INTEGRITY_VERIFICATION: 0.1,
			}
			outputWeights := map[core.EvidenceType]int{
				core.AIV_SECURE_BOOT:                          2,
				core.AIV_ACCESS_CONTROL:                       0,
				core.AIV_CONTROL_FLOW_INTEGRITY:               2,
				core.AIV_SECURE_OTA:                           0,
				core.AIV_APPLICATION_ISOLATION:                1,
				core.AIV_CONFIGURATION_INTEGRITY_VERIFICATION: 2,
			}
			designTimeTrustOpinion, _ := subjectivelogic.NewOpinion(0.1, 0.1, 0.8, 0.5)
			return quantifier(m, &designTimeTrustOpinion, existenceWeights, outputWeights)
		},
	},
}

var trustSources []core.EvidenceType

func init() {

	//Extract list of used trust sources from trustSourceQuantifierInstances
	evidenceMap := make(map[core.EvidenceType]bool)
	for _, quantifier := range trustSourceQuantifiers {
		for _, evidence := range quantifier.Evidence {
			evidenceMap[evidence] = true
		}
	}
	trustSources = make([]core.EvidenceType, len(evidenceMap))
	i := 0
	for k := range evidenceMap {
		trustSources[i] = k
		i++
	}
}

type TrustModelTemplate struct {
	name                   string
	version                string
	trustSourceQuantifiers []core.TrustSourceQuantifier
	description            string
	rTL1                   subjectivelogic.Opinion
	rTL2                   subjectivelogic.Opinion
}

func CreateTrustModelTemplate(name string, version string, description string) core.TrustModelTemplate {
	rtl1, _ := subjectivelogic.NewOpinion(0.7, 0.2, 0.1, 0.5)
	rtl2, _ := subjectivelogic.NewOpinion(0.65, 0.25, 0.1, 0.5)
	return TrustModelTemplate{
		name:                   name,
		version:                version,
		trustSourceQuantifiers: trustSourceQuantifiers,
		description:            description,
		rTL1:                   rtl1,
		rTL2:                   rtl2,
	}
}

func (tmt TrustModelTemplate) EvidenceTypes() []core.EvidenceType {
	return trustSources
}

func (tmt TrustModelTemplate) Version() string {
	return tmt.version
}

func (tmt TrustModelTemplate) TemplateName() string {
	return tmt.name
}

func (tmt TrustModelTemplate) Description() string {
	return tmt.description
}

func (tmt TrustModelTemplate) Spawn(params map[string]string, context core.TafContext, channels core.TafChannels) (core.TrustModelInstance, error) {

	omega1, _ := subjectivelogic.NewOpinion(0.2, 0.1, 0.7, 0.5)
	omega2, _ := subjectivelogic.NewOpinion(0.15, 0.15, 0.7, 0.5)

	//return nil, errors.New("Reason")

	return &TrustModelInstance{
		//		id:          tmt.TemplateName() + "@" + tmt.Version() + "-" + fmt.Sprintf("%000000d", rand.IntN(999999)),
		id:          tmt.TemplateName() + "@" + tmt.Version() + "-001",
		version:     0,
		template:    tmt,
		omega1:      omega1,
		omega2:      omega2,
		fingerprint: 0,
	}, nil
}

func (tmt TrustModelTemplate) TrustSourceQuantifiers() []core.TrustSourceQuantifier {
	return tmt.trustSourceQuantifiers
}
