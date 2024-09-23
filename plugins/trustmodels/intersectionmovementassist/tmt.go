package intersectionmovementassist

import (
	"fmt"
	"github.com/vs-uulm/go-taf/pkg/core"
)

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

func (t TrustModelTemplate) EvidenceTypes() []core.EvidenceType {
	return []core.EvidenceType{}
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

func (t TrustModelTemplate) TrustSourceQuantifiers() []core.TrustSourceQuantifier {
	return nil
}

func (t TrustModelTemplate) Type() core.TrustModelTemplateType {
	return core.VEHICLE_TRIGGERED_TRUST_MODEL
}

func (t TrustModelTemplate) OnNewVehicle(identifier string, params map[string]string) (core.TrustModelInstance, error) {
	return &TrustModelInstance{
		id:       identifier,
		version:  0,
		template: t,
		objects:  map[string]bool{},
	}, nil
}

func (t TrustModelTemplate) Identifier() string {
	return fmt.Sprintf("%s@%s", t.TemplateName(), t.Version())
}
