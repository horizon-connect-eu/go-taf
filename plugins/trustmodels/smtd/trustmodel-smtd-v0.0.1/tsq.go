package trustmodel_smtd_v0_0_1

import (
	"errors"
	"github.com/horizon-connect-eu/go-taf/pkg/core"
	"github.com/horizon-connect-eu/go-taf/pkg/trustsource"
	"github.com/vs-uulm/go-subjectivelogic/pkg/subjectivelogic"
	"log"
	"math"
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

var defaultMBDWeightsNoDetection = map[trustsource.MisbehaviorDetector]float64{
	trustsource.MBD_DIST_PLAU:                   1,
	trustsource.MBD_SPEE_PLAU:                   1,
	trustsource.MBD_SPEE_CONS:                   1,
	trustsource.MBD_POS_SPEE_CONS:               1,
	trustsource.MBD_KALMAN_POS_CONS:             2,
	trustsource.MBD_KALMAN_POS_SPEED_CONS_SPEED: 2,
	trustsource.MBD_KALMAN_POS_SPEED_CONS_POS:   2,
	trustsource.MBD_LOCAL_PERCEPTION_VERIF:      2,
}

var defaultMBDWeightsDetection = map[trustsource.MisbehaviorDetector]float64{
	trustsource.MBD_DIST_PLAU:                   2,
	trustsource.MBD_SPEE_PLAU:                   2,
	trustsource.MBD_SPEE_CONS:                   2,
	trustsource.MBD_POS_SPEE_CONS:               2,
	trustsource.MBD_KALMAN_POS_CONS:             1,
	trustsource.MBD_KALMAN_POS_SPEED_CONS_SPEED: 1,
	trustsource.MBD_KALMAN_POS_SPEED_CONS_POS:   1,
	trustsource.MBD_LOCAL_PERCEPTION_VERIF:      2,
}

func createTrustSourceQuantifiers(params map[string]string) ([]core.TrustSourceQuantifier, error) {
	mbdWeightsDetection := make(map[trustsource.MisbehaviorDetector]float64)

	for key, defaultValue := range defaultMBDWeightsDetection {
		mbdWeightsDetection[key] = defaultValue
	}

	mbdWeightsNoDetection := make(map[trustsource.MisbehaviorDetector]float64)

	for key, defaultValue := range defaultMBDWeightsNoDetection {
		mbdWeightsNoDetection[key] = defaultValue
	}

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
			if strings.Contains(key, "MBD_D") {
				detector := strings.SplitAfterN(key, "_", 3)[2]
				if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
					detectorType := trustsource.MisbehaviorDetectorByName("MBD_" + detector)
					if detectorType == trustsource.MBD_UNKNOWN {
						return nil, errors.New("Key" + key + "is not valid")
					} else {
						mbdWeightsDetection[detectorType] = floatValue
					}
				} else {
					return nil, errors.New("Key" + key + "is not a float value")
				}
			} else if strings.Contains(key, "MBD_ND") {
				detector := strings.SplitAfterN(key, "_", 3)[2]
				if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
					detectorType := trustsource.MisbehaviorDetectorByName("MBD_" + detector)
					if detectorType == trustsource.MBD_UNKNOWN {
						return nil, errors.New("Key" + key + "is not valid")
					} else {
						mbdWeightsNoDetection[detectorType] = floatValue
					}
				} else {
					return nil, errors.New("Key" + key + "is not a float value")
				}
			} else if strings.Contains(key, "TCH_EXISTENCE") {
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
		Trustor:     "V_ego",
		Trustee:     "V_*",
		Scope:       "C_*_*",
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

	mbdQuantifier := core.TrustSourceQuantifier{
		Trustor:     "V_ego",
		Trustee:     "C_*_*",
		Scope:       "C_*_*",
		TrustSource: core.MBD,
		Evidence:    []core.EvidenceType{core.MBD_MISBEHAVIOR_REPORT},
		Quantifier: func(m map[core.EvidenceType]interface{}) subjectivelogic.QueryableOpinion {
			binaryFormat := strconv.FormatInt(int64(m[core.MBD_MISBEHAVIOR_REPORT].(int)), 2)

			for i := len(binaryFormat); i < 8; i++ {
				binaryFormat = "0" + binaryFormat
			}

			sumWeights := 0.0
			sumBelief := 0.0
			sumDisbelief := 0.0

			for i := 0; i < 8; i++ {
				detector := trustsource.MisbehaviorDetector(7 - i)
				if string(binaryFormat[i]) == "0" {
					sumWeights = sumWeights + mbdWeightsNoDetection[detector]
					sumBelief = sumBelief + mbdWeightsNoDetection[detector]
				} else {
					sumWeights = sumWeights + mbdWeightsDetection[detector]
					sumDisbelief = sumDisbelief + mbdWeightsDetection[detector]
				}
			}

			exponentialValue := -math.Pow(1.3, -float64(sumWeights)) + 1
			belief := (sumBelief / sumWeights) * exponentialValue
			disbelief := (sumDisbelief / sumWeights) * exponentialValue
			//uncertainty := 1 - exponentialValue

			sl, _ := subjectivelogic.NewOpinion(belief, disbelief, 1-belief-disbelief, 0.5)

			return &sl
		},
	}

	//TODO: return TSQs with local mbdWeightsDetection, etc.

	return []core.TrustSourceQuantifier{tchQuantifier, mbdQuantifier}, nil
}

func init() {
	//TODO: Validate weights here instead of inside quantifier function
}
