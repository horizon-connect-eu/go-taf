package trustmodel_ima_standalone_v0_0_2

import (
	"fmt"
	"github.com/horizon-connect-eu/go-taf/pkg/core"
	"github.com/vs-uulm/go-subjectivelogic/pkg/subjectivelogic"
	"strconv"
)

var FullBelief, _ = subjectivelogic.NewOpinion(1, 0, 0, 0.5)
var FullUncertainty, _ = subjectivelogic.NewOpinion(0, 0, 1, 0.5)
var RTL, _ = subjectivelogic.NewOpinion(1, 0, 0, 0.5)

var DEFAULT_MBD_EWMA_ALPHA = 1.00

type TrustModelTemplate struct {
	name          string
	version       string
	evidenceTypes []core.EvidenceType
	params        map[string]string
}

func CreateTrustModelTemplate(name string, version string) core.TrustModelTemplate {

	//Extract the list of used trust sources from TrustSourceQuantifiers
	tsqs, _ := createTrustSourceQuantifiers(nil)
	evidenceMap := make(map[core.EvidenceType]bool)
	for _, quantifier := range tsqs {
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
	tsqs, err := createTrustSourceQuantifiers(params)
	if err != nil {
		return nil, nil, nil, err
	} else {
		spawner := NewDynamicTrustModelTemplateSpawner(t, params)
		return tsqs, nil, spawner, nil
	}
}

func (t TrustModelTemplate) Description() string {
	return "IMA Trust Model, standalone variant. This trust model supports configurable exponentially weighted averaging of ATOs from trust source MBD."
}

func (t TrustModelTemplate) Type() core.TrustModelTemplateType {
	return core.VEHICLE_TRIGGERED_TRUST_MODEL
}

func (t TrustModelTemplate) Identifier() string {
	return fmt.Sprintf("%s@%s", t.TemplateName(), t.Version())
}

func (t TrustModelTemplate) EvidenceTypes() []core.EvidenceType {
	return t.evidenceTypes
}

func (tmt TrustModelTemplate) SigningHash() string {
	return SigningHash
}

type DynamicTrustModelTemplateSpawner struct {
	template TrustModelTemplate
	params   map[string]string
}

func NewDynamicTrustModelTemplateSpawner(template TrustModelTemplate, params map[string]string) DynamicTrustModelTemplateSpawner {
	return DynamicTrustModelTemplateSpawner{
		template: template,
		params:   params,
	}
}

func (t DynamicTrustModelTemplateSpawner) OnNewTrustee(identifier string, params map[string]string) (core.TrustModelInstance, error) {
	return nil, nil
}

func (t DynamicTrustModelTemplateSpawner) OnNewVehicle(identifier string, params map[string]string) (core.TrustModelInstance, error) {
	initialParams := t.params
	newParams := params
	params = map[string]string{}
	//add parameters set at Spawn() call
	if initialParams != nil {
		for key, value := range initialParams {
			params[key] = value
		}
	}
	//add/overwrite parameters set at OnNewVehicle() call
	if newParams != nil {
		for key, value := range newParams {
			params[key] = value
		}
	}

	ewmaAlpha := DEFAULT_MBD_EWMA_ALPHA
	alpha, exists := params["MBD_EWMA_ALPHA"]
	if exists {
		if parsedAlpha, err := strconv.ParseFloat(alpha, 64); err == nil && parsedAlpha <= 1 && parsedAlpha > 0 {
			ewmaAlpha = parsedAlpha
		}
	}

	return &TrustModelInstance{
		id:        identifier,
		version:   0,
		template:  t.template,
		objects:   map[string]subjectivelogic.QueryableOpinion{},
		staticRTL: &RTL,
		ewmaAlpha: ewmaAlpha,
	}, nil
}
