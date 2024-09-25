package intersectionmovementassist

import (
	"fmt"
	"github.com/vs-uulm/go-subjectivelogic/pkg/subjectivelogic"
	"github.com/vs-uulm/go-taf/pkg/core"
	"log"
)

var FullBelief, _ = subjectivelogic.NewOpinion(1, 0, 0, 0.5)
var RTL, _ = subjectivelogic.NewOpinion(1, 0, 0, 0.5)
var iDontKnow, _ = subjectivelogic.NewOpinion(0, 1, 0, 0.5)

var tchExistenceWeights = map[core.EvidenceType]float64{
	core.TCH_SECURE_BOOT:                          0.24,
	core.TCH_ACCESS_CONTROL:                       0.16,
	core.TCH_CONTROL_FLOW_INTEGRITY:               0.08,
	core.TCH_SECURE_OTA:                           0.08,
	core.TCH_APPLICATION_ISOLATION:                0.16,
	core.TCH_CONFIGURATION_INTEGRITY_VERIFICATION: 0.24,
}

var tchOutputWeights = map[core.EvidenceType]int{
	core.TCH_SECURE_BOOT:                          2,
	core.TCH_ACCESS_CONTROL:                       1,
	core.TCH_CONTROL_FLOW_INTEGRITY:               2,
	core.TCH_SECURE_OTA:                           0,
	core.TCH_APPLICATION_ISOLATION:                0,
	core.TCH_CONFIGURATION_INTEGRITY_VERIFICATION: 2,
}

var trustSourceQuantifiers = []core.TrustSourceQuantifier{
	{
		Trustor:     "V_ego",
		Trustee:     "V_*",
		Scope:       "C_*_*",
		TrustSource: core.TCH,
		Evidence:    []core.EvidenceType{core.TCH_SECURE_BOOT, core.TCH_SECURE_OTA, core.TCH_ACCESS_CONTROL, core.TCH_APPLICATION_ISOLATION, core.TCH_CONTROL_FLOW_INTEGRITY, core.TCH_CONFIGURATION_INTEGRITY_VERIFICATION},
		Quantifier: func(m map[core.EvidenceType]int) subjectivelogic.QueryableOpinion {
			sl, _ := subjectivelogic.NewOpinion(.0, .0, 1.0, 0.5)

			fmt.Printf("%+v ", tchExistenceWeights)

			var sum = 0.0
			for _, val := range tchExistenceWeights {
				sum += val
			}

			if sum > 1.0 { // sum of existence weights is not allowed to exceed 1.0
				log.Fatalf("Sum existence weights of the TCH trust source exceeds 1.0")
			}

			belief := 0.0
			disbelief := 0.0
			uncertainty := 1.0

			for control, appraisal := range m {
				delta, ok := tchExistenceWeights[control]

				if ok { // Only if control is one of the forseen controls, belief and disbelief will be adjusted
					if appraisal == -1 { // control not implemented
						disbelief = disbelief + delta
						uncertainty = uncertainty - delta
					} else if appraisal == 0 {
						if tchOutputWeights[control] == 0 { // still add belief
							belief = belief + delta
							uncertainty = uncertainty - delta
						} else if tchOutputWeights[control] == 1 { // add disbelief
							disbelief = disbelief + delta
							uncertainty = uncertainty - delta
						} else if tchOutputWeights[control] == 2 { // complete disbelief
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
			}

			sl.Modify(belief, disbelief, uncertainty, sl.BaseRate())

			return &sl
		},
	},
	{
		Trustor:     "V_ego",
		Trustee:     "C_*_*",
		Scope:       "C_*_*",
		TrustSource: core.MBD,
		Evidence:    []core.EvidenceType{core.MBD_MISBEHAVIOR_REPORT},
		Quantifier: func(m map[core.EvidenceType]int) subjectivelogic.QueryableOpinion {
			//TODO: implement
			return &iDontKnow
		},
	},
}

type TrustModelTemplate struct {
	name    string
	version string
}

func CreateTrustModelTemplate(name string, version string) core.TrustModelTemplate {
	return TrustModelTemplate{
		name:    name,
		version: version,
	}
}

func (t TrustModelTemplate) Version() string {
	return t.version
}

func (t TrustModelTemplate) TemplateName() string {
	return t.name
}

func (t TrustModelTemplate) Spawn(params map[string]string, context core.TafContext) (core.TrustModelInstance, core.DynamicTrustModelInstanceSpawner, error) {
	return nil, t, nil
}

func (t TrustModelTemplate) Description() string {
	return "TODO: Add description of trust model"
}

func (t TrustModelTemplate) Type() core.TrustModelTemplateType {
	return core.VEHICLE_TRIGGERED_TRUST_MODEL
}

func (t TrustModelTemplate) OnNewVehicle(identifier string, params map[string]string) (core.TrustModelInstance, error) {
	return &TrustModelInstance{
		id:        identifier,
		version:   0,
		template:  t,
		objects:   map[string]bool{},
		staticRTL: &RTL,
	}, nil
}

func (t TrustModelTemplate) Identifier() string {
	return fmt.Sprintf("%s@%s", t.TemplateName(), t.Version())
}

func (t TrustModelTemplate) EvidenceTypes() []core.EvidenceType {
	// TODO: implement
	return []core.EvidenceType{}
}

func (t TrustModelTemplate) TrustSourceQuantifiers() []core.TrustSourceQuantifier {
	return trustSourceQuantifiers
}
