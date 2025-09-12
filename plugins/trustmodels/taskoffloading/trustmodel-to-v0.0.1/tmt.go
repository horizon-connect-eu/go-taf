package trustmodel_to_v0_0_1

import (
	"fmt"
	"github.com/horizon-connect-eu/go-taf/pkg/core"
	"github.com/vs-uulm/go-subjectivelogic/pkg/subjectivelogic"
)

var FullUncertainty, _ = subjectivelogic.NewOpinion(0, 0, 1, 0.5)

var DefaultRTL, _ = subjectivelogic.NewOpinion(0.7, 0.2, 0.1, 0.5)

type TrustModelTemplate struct {
	name          string
	version       string
	evidenceTypes []core.EvidenceType
	params        map[string]string
}

func CreateTrustModelTemplate(name string, version string) core.TrustModelTemplate {

	//Extract list of used trust sources from TrustSourceQuantifiers
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

func (t TrustModelTemplate) Description() string {
	return "Task Offloading Trust Model"
}

func (tmt TrustModelTemplate) Type() core.TrustModelTemplateType {
	return core.TRUSTEE_TRIGGERED_TRUST_MODEL
}

func (t TrustModelTemplate) OnNewVehicle(identifier string, params map[string]string) (core.TrustModelInstance, error) {
	return nil, nil
}

func (t TrustModelTemplate) Version() string {
	return t.version
}

func (t TrustModelTemplate) TemplateName() string {
	return t.name
}

func (tmt TrustModelTemplate) Identifier() string {
	return fmt.Sprintf("%s@%s", tmt.TemplateName(), tmt.Version())
}

func (t TrustModelTemplate) EvidenceTypes() []core.EvidenceType {
	return t.evidenceTypes
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

func (tmt TrustModelTemplate) SigningHash() string {
	return SigningHash
}

func (t TrustModelTemplate) OnNewTrustee(identifier string, params map[string]string) (core.TrustModelInstance, error) {
	initialParams := t.params
	newParams := params
	params = map[string]string{}
	//add parameters set at Spawn() call
	if initialParams != nil {
		for key, value := range initialParams {
			params[key] = value
		}
	}
	//add/overwrite parameters set at OnNewTrustee() call
	if newParams != nil {
		for key, value := range newParams {
			params[key] = value
		}
	}

	return &TrustModelInstance{
		id:       identifier,
		version:  0,
		template: t,
		omega:    FullUncertainty,
	}, nil
}
