package vcm_v0_0_1

import (
	"errors"
	"fmt"
	"github.com/vs-uulm/go-subjectivelogic/pkg/subjectivelogic"
	"github.com/vs-uulm/go-taf/pkg/core"
	"math/rand/v2"
	"strconv"
	"strings"
)

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

func getExistenceWeightsFromInit(params map[string]string, key string) (float64, error) {
	if strValue, found := params[key]; found {
		if floatValue, err := strconv.ParseFloat(strValue, 64); err == nil {
			return floatValue, nil
		} else {
			return -1.0, errors.New("Key" + key + "is not a float value")
		}
	} else {
		return -1.0, errors.New("Key" + key + "not provided")
	}
}

func getOutputWeightsFromInit(params map[string]string, key string) (int, error) {
	if strValue, found := params[key]; found {
		if intValue, err := strconv.Atoi(strValue); err == nil {
			return intValue, nil
		} else {
			return -1, errors.New("Key" + key + "is not an int value")
		}
	} else {
		return -1, errors.New("Key" + key + "not provided")
	}
}

func getOpinionFromInit(params map[string]string, opinionKey string) (subjectivelogic.Opinion, error) {
	belief, err := getExistenceWeightsFromInit(params, opinionKey+"_BELIEF")
	if err != nil {
		return subjectivelogic.Opinion{}, err
	}

	disbelief, err := getExistenceWeightsFromInit(params, opinionKey+"_DISBELIEF")
	if err != nil {
		return subjectivelogic.Opinion{}, err
	}

	uncertainty, err := getExistenceWeightsFromInit(params, opinionKey+"_UNCERTAINTY")
	if err != nil {
		return subjectivelogic.Opinion{}, err
	}

	baserate, err := getExistenceWeightsFromInit(params, opinionKey+"_BASERATE")
	if err != nil {
		return subjectivelogic.Opinion{}, err
	}

	opinion, err := subjectivelogic.NewOpinion(belief, disbelief, uncertainty, baserate)

	return opinion, err
}

func checkSetParameters(params map[string]string) map[string]bool {
	setParams := make(map[string]bool)

	for k := range params {
		if strings.HasPrefix(k, "VC1_EXISTENCE") {
			setParams["VC1_EXISTENCE"] = true
		} else if strings.HasPrefix(k, "VC2_EXISTENCE") {
			setParams["VC2_EXISTENCE"] = true
		} else if strings.HasPrefix(k, "VC1_OUTPUT") {
			setParams["VC1_OUTPUT"] = true
		} else if strings.HasPrefix(k, "VC2_OUTPUT") {
			setParams["VC2_OUTPUT"] = true
		} else if strings.HasPrefix(k, "VC1_DTI") {
			setParams["VC1_DTI"] = true
		} else if strings.HasPrefix(k, "VC2_DTI") {
			setParams["VC2_DTI"] = true
		} else if strings.HasPrefix(k, "VC1_RTL") {
			setParams["VC1_RTL"] = true
		} else if strings.HasPrefix(k, "VC2_RTL") {
			setParams["VC2_RTL"] = true
		}
	}

	return setParams
}

func (tmt TrustModelTemplate) Spawn(params map[string]string, context core.TafContext) ([]core.TrustSourceQuantifier, core.TrustModelInstance, core.DynamicTrustModelInstanceSpawner, error) {
	setParams := checkSetParameters(params)

	omega1, _ := subjectivelogic.NewOpinion(0.0, 0.0, 1.0, 0.5)
	omega2, _ := subjectivelogic.NewOpinion(0.0, 0.0, 1.0, 0.5)

	if len(params) > 0 {
		// get existence parameters for VC1
		sum := 0.0

		if _, found := setParams["VC1_EXISTENCE"]; found {
			for _, typeEvidence := range tmt.trustSourceQuantifiers[0].Evidence {
				value, err := getExistenceWeightsFromInit(params, "VC1_EXISTENCE_"+typeEvidence.String())
				if err != nil {
					return nil, nil, nil, err
				}
				vc1ExistenceWeights[typeEvidence] = value

				sum = sum + value
			}

			if sum > 1 {
				return nil, nil, nil, errors.New("Values for existence weights of VC1 sum up to more than 1")
			}

		}

		// get existence parameters for VC2
		sum = 0.0

		if _, found := setParams["VC2_EXISTENCE"]; found {
			for _, typeEvidence := range tmt.trustSourceQuantifiers[1].Evidence {
				value, err := getExistenceWeightsFromInit(params, "VC2_EXISTENCE_"+typeEvidence.String())
				if err != nil {
					return nil, nil, nil, err
				}
				vc2ExistenceWeights[typeEvidence] = value

				sum = sum + value
			}

			if sum > 1 {
				return nil, nil, nil, errors.New("Values for existence weights of VC2 sum up to more than 1")
			}
		}

		// get output parameters for VC1
		if _, found := setParams["VC1_OUTPUT"]; found {
			for _, typeEvidence := range tmt.trustSourceQuantifiers[0].Evidence {
				value, err := getOutputWeightsFromInit(params, "VC1_OUTPUT_"+typeEvidence.String())
				if err != nil {
					return nil, nil, nil, err
				}
				vc1OutputWeights[typeEvidence] = value

				if value < 0 || value > 2 {
					return nil, nil, nil, errors.New("Invalid value for VC1_OUTPUT_" + typeEvidence.String() + "- value has to be between 0 and 2")
				}
			}
		}

		// get output parameters for VC2
		if _, found := setParams["VC2_OUTPUT"]; found {
			for _, typeEvidence := range tmt.trustSourceQuantifiers[1].Evidence {
				value, err := getOutputWeightsFromInit(params, "VC2_OUTPUT_"+typeEvidence.String())
				if err != nil {
					return nil, nil, nil, err
				}
				vc2OutputWeights[typeEvidence] = value

				if value < 0 || value > 2 {
					return nil, nil, nil, errors.New("Invalid value for VC2_OUTPUT_" + typeEvidence.String() + "- value has to be between 0 and 2")
				}
			}
		}

		// get DTI for VC1
		if _, found := setParams["VC1_DTI"]; found {
			err := errors.New("")
			vc1DTI, err = getOpinionFromInit(params, "VC1_DTI")
			if err != nil {
				return nil, nil, nil, err
			}
		}

		// get DTI for VC1
		if _, found := setParams["VC2_DTI"]; found {
			err := errors.New("")
			vc2DTI, err = getOpinionFromInit(params, "VC2_DTI")
			if err != nil {
				return nil, nil, nil, err
			}
		}

		// get RTL for VC1
		if _, found := setParams["VC1_RTL"]; found {
			err := errors.New("")
			tmt.rTL1, err = getOpinionFromInit(params, "VC1_RTL")
			if err != nil {
				return nil, nil, nil, err
			}
		}

		// get DTI for VC1
		if _, found := setParams["VC2_RTL"]; found {
			err := errors.New("")
			tmt.rTL2, err = getOpinionFromInit(params, "VC2_RTL")
			if err != nil {
				return nil, nil, nil, err
			}
		}

	}

	return tmt.trustSourceQuantifiers, &TrustModelInstance{
		id:          fmt.Sprintf("%000000d", rand.IntN(999999)),
		version:     0,
		template:    tmt,
		omega1:      omega1,
		omega2:      omega2,
		fingerprint: rand.Uint32N(999999999),
	}, nil, nil
}

func (tmt TrustModelTemplate) TrustSourceQuantifiers() []core.TrustSourceQuantifier {
	return tmt.trustSourceQuantifiers
}

func (tmt TrustModelTemplate) Identifier() string {
	return fmt.Sprintf("%s@%s", tmt.TemplateName(), tmt.Version())
}

func (tmt TrustModelTemplate) Type() core.TrustModelTemplateType {
	return core.STATIC_TRUST_MODEL
}
