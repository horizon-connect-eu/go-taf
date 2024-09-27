package intersectionmovementassist

import (
	"fmt"
	"github.com/vs-uulm/go-subjectivelogic/pkg/subjectivelogic"
	"github.com/vs-uulm/go-taf/pkg/core"
)

var FullBelief, _ = subjectivelogic.NewOpinion(1, 0, 0, 0.5)
var FullUncertainty, _ = subjectivelogic.NewOpinion(0, 0, 1, 0.5)
var RTL, _ = subjectivelogic.NewOpinion(1, 0, 0, 0.5)

type TrustModelTemplate struct {
	name          string
	version       string
	evidenceTypes []core.EvidenceType
	params        map[string]string
}

func CreateTrustModelTemplate(name string, version string) core.TrustModelTemplate {

	//Extract list of used trust sources from trustSourceQuantifierInstances
	evidenceMap := make(map[core.EvidenceType]bool)
	for _, quantifier := range trustSourceQuantifiers {
		for _, evidence := range quantifier.Evidence {
			evidenceMap[evidence] = true
		}
	}
	evidenceTypes := make([]core.EvidenceType, len(evidenceMap))
	i := 0
	for k := range evidenceMap {
		evidenceTypes[i] = k
		i++
	}

	return TrustModelTemplate{
		name:          name,
		version:       version,
		evidenceTypes: evidenceTypes,
	}
}

func (t TrustModelTemplate) Version() string {
	return t.version
}

func (t TrustModelTemplate) TemplateName() string {
	return t.name
}

func (t TrustModelTemplate) Spawn(params map[string]string, context core.TafContext) ([]core.TrustSourceQuantifier, core.TrustModelInstance, core.DynamicTrustModelInstanceSpawner, error) {
	t.params = params
	tsqs, err := createTrustSourceQuantifiers(params)
	if err != nil {
		return nil, nil, nil, err
	} else {
		return tsqs, nil, t, nil
	}
}

func (t TrustModelTemplate) Description() string {
	return "TODO: Add description of trust model"
}

func (t TrustModelTemplate) Type() core.TrustModelTemplateType {
	return core.VEHICLE_TRIGGERED_TRUST_MODEL
}

func (t TrustModelTemplate) OnNewVehicle(identifier string, params map[string]string) (core.TrustModelInstance, error) {
	initialParams := t.params
	newParams := params
	params = map[string]string{}
	//add parameters set at Spawn()
	if initialParams != nil {
		for key, value := range initialParams {
			params[key] = value
		}
	}
	//add parameters set at OnNewVehicle()
	if newParams != nil {
		for key, value := range newParams {
			params[key] = value
		}
	}

	//TODO: use params to set weights at runtime

	return &TrustModelInstance{
		id:        identifier,
		version:   0,
		template:  t,
		objects:   map[string]subjectivelogic.QueryableOpinion{},
		staticRTL: &RTL,
	}, nil
}

func (t TrustModelTemplate) Identifier() string {
	return fmt.Sprintf("%s@%s", t.TemplateName(), t.Version())
}

func (t TrustModelTemplate) EvidenceTypes() []core.EvidenceType {
	// TODO: implement
	return t.evidenceTypes
}

func (t TrustModelTemplate) TrustSourceQuantifiers() []core.TrustSourceQuantifier {
	//return trustSourceQuantifiers
	return []core.TrustSourceQuantifier{}
}
