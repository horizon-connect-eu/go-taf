package trustmodel_to_v0_0_1

import (
	"errors"
	"github.com/horizon-connect-eu/go-taf/pkg/core"
	"github.com/vs-uulm/go-subjectivelogic/pkg/subjectivelogic"
	"log"
	"strconv"
	"strings"
)

var defaultTCHExistenceWeights = map[core.EvidenceType]float64{
	core.TCH_SECURE_BOOT:                          0.24,
	core.TCH_ACCESS_CONTROL:                       0.16,
	core.TCH_CONTROL_FLOW_INTEGRITY:               0.08,
	core.TCH_SECURE_OTA:                           0.08,
	core.TCH_APPLICATION_ISOLATION:                0.16,
	core.TCH_CONFIGURATION_INTEGRITY_VERIFICATION: 0.24,
}

var defaultTCHOutputWeights = map[core.EvidenceType]float64{
	core.TCH_SECURE_BOOT:                          2,
	core.TCH_ACCESS_CONTROL:                       1,
	core.TCH_CONTROL_FLOW_INTEGRITY:               2,
	core.TCH_SECURE_OTA:                           0,
	core.TCH_APPLICATION_ISOLATION:                0,
	core.TCH_CONFIGURATION_INTEGRITY_VERIFICATION: 2,
}

func createTrustSourceQuantifiers(params map[string]string) ([]core.TrustSourceQuantifier, error) {

	tchExistenceWeights := make(map[core.EvidenceType]float64)

	for key, defaultValue := range defaultTCHExistenceWeights {
		tchExistenceWeights[key] = defaultValue
	}

	tchOutputWeights := make(map[core.EvidenceType]float64)

	for key, defaultValue := range defaultTCHOutputWeights {
		tchOutputWeights[key] = defaultValue
	}

	if params != nil {
		//TODO: update  with params

		for key, value := range params {
			if strings.Contains(key, "TCH_EXISTENCE") {
				control := strings.SplitAfterN(key, "_", 3)[2]
				if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
					controlType := core.EvidenceTypeBySourceAndName(core.TCH, control)
					if controlType == core.UNKNOWN {
						return nil, errors.New("Key" + key + "is not valid")
					} else {
						tchExistenceWeights[controlType] = floatValue
					}
				} else {
					return nil, errors.New("Key" + key + "is not a float value")
				}
			} else if strings.Contains(key, "TCH_OUTPUT") {
				control := strings.SplitAfterN(key, "_", 3)[2]
				if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
					controlType := core.EvidenceTypeBySourceAndName(core.TCH, control)
					if controlType == core.UNKNOWN {
						return nil, errors.New("Key" + key + "is not valid")
					} else {
						tchOutputWeights[controlType] = floatValue
					}
				} else {
					return nil, errors.New("Key" + key + "is not a float value")
				}
			}
		}
	}

	// normalization of TCH_EXISTENCE weights in case the sum is more than 1.0
	sum := 0.0
	for _, value := range tchExistenceWeights {
		sum = sum + value
	}

	if sum > 1.0 {
		for key, value := range tchExistenceWeights {
			tchExistenceWeights[key] = value / sum
		}
	}

	tchQuantifier := core.TrustSourceQuantifier{
		Trustor:     "MEC",
		Trustee:     "vehicle_*",
		Scope:       "vehicle_*",
		TrustSource: core.TCH,
		Evidence:    []core.EvidenceType{core.TCH_SECURE_BOOT, core.TCH_SECURE_OTA, core.TCH_ACCESS_CONTROL, core.TCH_APPLICATION_ISOLATION, core.TCH_CONTROL_FLOW_INTEGRITY, core.TCH_CONFIGURATION_INTEGRITY_VERIFICATION},
		Quantifier: func(m map[core.EvidenceType]interface{}) subjectivelogic.QueryableOpinion {

			var sum = 0.0
			for _, val := range tchExistenceWeights {
				sum += val
			}

			if sum > 1.0 { // sum of existence weights is not allowed to exceed 1.0
				log.Fatalf("Sum existence weights of the TCH trust source exceeds 1.0")
			}

			belief := 0.0
			disbelief := 0.0
			//uncertainty := 1.0

			for control, rawAppraisal := range m {
				appraisal := rawAppraisal.(int)
				delta, ok := tchExistenceWeights[control]

				if ok { // Only if control is one of the foreseen controls, belief and disbelief will be adjusted
					if appraisal == -1 { // control not implemented
						disbelief = disbelief + delta
						//uncertainty = uncertainty - delta
					} else if appraisal == 0 {
						if tchOutputWeights[control] == 0 { // still add belief
							belief = belief + delta
							//uncertainty = uncertainty - delta
						} else if tchOutputWeights[control] == 1 { // add disbelief
							disbelief = disbelief + delta
							//uncertainty = uncertainty - delta
						} else if tchOutputWeights[control] == 2 { // complete disbelief
							belief = 0.0
							disbelief = 1.0
							//uncertainty = 0.0
							break // complete disbelief because negative evidence of critical securityControl
						} else {
							// Invalid weight
							// TODO: Error handling
						}
					} else if appraisal == 1 {
						belief = belief + delta
						//uncertainty = uncertainty - delta
					} else {
						// No evidence for the control, e.g. appraisal -2 or no evidence received -> Results in higher uncertainty
					}
				}
			}
			sl, _ := subjectivelogic.NewOpinion(belief, disbelief, 1-belief-disbelief, 0.5)

			return &sl
		},
	}

	return []core.TrustSourceQuantifier{tchQuantifier}, nil
}

func init() {
	//TODO: Validate weights here instead of inside quantifier function
}
