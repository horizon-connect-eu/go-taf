package intersectionmovementassist

import (
	"errors"
	"fmt"
	"github.com/vs-uulm/go-subjectivelogic/pkg/subjectivelogic"
	"github.com/vs-uulm/go-taf/pkg/core"
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

var defaultMBDWeightsNoDetection = map[core.MisbehaviorDetector]float64{
	core.MBD_DIST_PLAU:                   1,
	core.MBD_SPEE_PLAU:                   1,
	core.MBD_SPEE_CONS:                   1,
	core.MBD_POS_SPEE_CONS:               1,
	core.MBD_KALMAN_POS_CONS:             2,
	core.MBD_KALMAN_POS_SPEED_CONS_SPEED: 2,
	core.MBD_KALMAN_POS_SPEED_CONS_POS:   2,
	core.MBD_LOCAL_PERCEPTION_VERIF:      2,
}

var defaultMBDWeightsDetection = map[core.MisbehaviorDetector]float64{
	core.MBD_DIST_PLAU:                   2,
	core.MBD_SPEE_PLAU:                   2,
	core.MBD_SPEE_CONS:                   2,
	core.MBD_POS_SPEE_CONS:               2,
	core.MBD_KALMAN_POS_CONS:             1,
	core.MBD_KALMAN_POS_SPEED_CONS_SPEED: 1,
	core.MBD_KALMAN_POS_SPEED_CONS_POS:   1,
	core.MBD_LOCAL_PERCEPTION_VERIF:      2,
}

func createTrustSourceQuantifiers(params map[string]string) ([]core.TrustSourceQuantifier, error) {
	mbdWeightsDetection := make(map[core.MisbehaviorDetector]float64)

	for key, defaultValue := range defaultMBDWeightsDetection {
		mbdWeightsDetection[key] = defaultValue
	}

	mbdWeightsNoDetection := make(map[core.MisbehaviorDetector]float64)

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
				detector := strings.SplitAfterN(key, "_", 2)[2]
				if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
					detectorType := core.MisbehaviorDetectorByName("MBD_" + detector)
					if detectorType == core.MBD_UNKNOWN {
						return nil, errors.New("Key" + key + "is not valid")
					} else {
						mbdWeightsDetection[detectorType] = floatValue
					}
				} else {
					return nil, errors.New("Key" + key + "is not a float value")
				}
			} else if strings.Contains(key, "MBD_ND") {
				detector := strings.SplitAfterN(key, "_", 2)[2]
				if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
					detectorType := core.MisbehaviorDetectorByName("MBD_" + detector)
					if detectorType == core.MBD_UNKNOWN {
						return nil, errors.New("Key" + key + "is not valid")
					} else {
						mbdWeightsNoDetection[detectorType] = floatValue
					}
				} else {
					return nil, errors.New("Key" + key + "is not a float value")
				}
			} else if strings.Contains(key, "TCH_EXISTENCE") {
				control := strings.SplitAfterN(key, "_", 2)[2]
				if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
					controlType := core.EvidenceTypeByName("TCH_" + control)
					if controlType == core.UNKNOWN {
						return nil, errors.New("Key" + key + "is not valid")
					} else {
						tchExistenceWeights[controlType] = floatValue
					}
				} else {
					return nil, errors.New("Key" + key + "is not a float value")
				}
			} else if strings.Contains(key, "TCH_OUTPUT") {
				control := strings.SplitAfterN(key, "_", 2)[2]
				if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
					controlType := core.EvidenceTypeByName("TCH_" + control)
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

	tchQuantifier := core.TrustSourceQuantifier{
		Trustor:     "V_ego",
		Trustee:     "V_*",
		Scope:       "C_*_*",
		TrustSource: core.TCH,
		Evidence:    []core.EvidenceType{core.TCH_SECURE_BOOT, core.TCH_SECURE_OTA, core.TCH_ACCESS_CONTROL, core.TCH_APPLICATION_ISOLATION, core.TCH_CONTROL_FLOW_INTEGRITY, core.TCH_CONFIGURATION_INTEGRITY_VERIFICATION},
		Quantifier: func(m map[core.EvidenceType]int) subjectivelogic.QueryableOpinion {
			sl, _ := subjectivelogic.NewOpinion(.0, .0, 1.0, 0.5)

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

			for control, appraisal := range m {
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

			sl.Modify(belief, disbelief, 1-belief-disbelief, sl.BaseRate())

			return &sl
		},
	}

	mbdQuantifier := core.TrustSourceQuantifier{
		Trustor:     "V_ego",
		Trustee:     "C_*_*",
		Scope:       "C_*_*",
		TrustSource: core.MBD,
		Evidence:    []core.EvidenceType{core.MBD_MISBEHAVIOR_REPORT},
		Quantifier: func(m map[core.EvidenceType]int) subjectivelogic.QueryableOpinion {
			binaryFormat := strconv.FormatInt(int64(m[core.MBD_MISBEHAVIOR_REPORT]), 2)

			for i := len(binaryFormat); i < 8; i++ {
				binaryFormat = "0" + binaryFormat
			}

			sumWeights := 0.0
			sumBelief := 0.0
			sumDisbelief := 0.0

			for i := 0; i < 8; i++ {
				detector := core.MisbehaviorDetector(7 - i)
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

var trustSourceQuantifiers = []core.TrustSourceQuantifier{
	{
		Trustor:     "V_ego",
		Trustee:     "V_*",
		Scope:       "C_*_*",
		TrustSource: core.TCH,
		Evidence:    []core.EvidenceType{core.TCH_SECURE_BOOT, core.TCH_SECURE_OTA, core.TCH_ACCESS_CONTROL, core.TCH_APPLICATION_ISOLATION, core.TCH_CONTROL_FLOW_INTEGRITY, core.TCH_CONFIGURATION_INTEGRITY_VERIFICATION},
		Quantifier: func(m map[core.EvidenceType]int) subjectivelogic.QueryableOpinion {
			sl, _ := subjectivelogic.NewOpinion(.0, .0, 1.0, 0.5)

			fmt.Printf("%+v ", defaultTCHExistenceWeights)

			var sum = 0.0
			for _, val := range defaultTCHExistenceWeights {
				sum += val
			}

			if sum > 1.0 { // sum of existence weights is not allowed to exceed 1.0
				log.Fatalf("Sum existence weights of the TCH trust source exceeds 1.0")
			}

			belief := 0.0
			disbelief := 0.0
			//uncertainty := 1.0

			for control, appraisal := range m {
				delta, ok := defaultTCHExistenceWeights[control]

				if ok { // Only if control is one of the foreseen controls, belief and disbelief will be adjusted
					if appraisal == -1 { // control not implemented
						disbelief = disbelief + delta
						//uncertainty = uncertainty - delta
					} else if appraisal == 0 {
						if defaultTCHOutputWeights[control] == 0 { // still add belief
							belief = belief + delta
							//uncertainty = uncertainty - delta
						} else if defaultTCHOutputWeights[control] == 1 { // add disbelief
							disbelief = disbelief + delta
							//uncertainty = uncertainty - delta
						} else if defaultTCHOutputWeights[control] == 2 { // complete disbelief
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

			sl.Modify(belief, disbelief, 1-belief-disbelief, sl.BaseRate())

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
			binaryFormat := strconv.FormatInt(int64(m[core.MBD_MISBEHAVIOR_REPORT]), 2)

			for i := len(binaryFormat); i < 8; i++ {
				binaryFormat = "0" + binaryFormat
			}

			sumWeights := 0.0
			sumBelief := 0.0
			sumDisbelief := 0.0

			for i := 0; i < 8; i++ {
				detector := core.MisbehaviorDetector(7 - i)
				if string(binaryFormat[i]) == "0" {
					sumWeights = sumWeights + defaultMBDWeightsNoDetection[detector]
					sumBelief = sumBelief + defaultMBDWeightsNoDetection[detector]
				} else {
					sumWeights = sumWeights + defaultMBDWeightsDetection[detector]
					sumDisbelief = sumDisbelief + defaultMBDWeightsDetection[detector]
				}
			}

			exponentialValue := -math.Pow(1.3, -float64(sumWeights)) + 1
			belief := (sumBelief / sumWeights) * exponentialValue
			disbelief := (sumDisbelief / sumWeights) * exponentialValue
			//uncertainty := 1 - exponentialValue

			sl, _ := subjectivelogic.NewOpinion(belief, disbelief, 1-belief-disbelief, 0.5)

			return &sl
		},
	},
}

func init() {
	//TODO: Validate weights here instead of inside quantifier function
}
