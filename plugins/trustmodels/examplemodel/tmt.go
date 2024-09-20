package examplemodel

import (
	"fmt"
	"github.com/vs-uulm/go-taf/pkg/core"
	"math/rand/v2"
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
	return &TrustModelInstance{
		id:       fmt.Sprintf("%000000d", rand.IntN(999999)),
		version:  0,
		template: t,
	}, nil, nil
}

func (t TrustModelTemplate) Description() string {
	return "TODO: Add description of trust model"
}

func (t TrustModelTemplate) TrustSourceQuantifiers() []core.TrustSourceQuantifier {
	return nil
}

func (tmt TrustModelTemplate) Type() core.TrustModelTemplateType {
	return core.STATIC_TRUST_MODEL
}

func (tmt TrustModelTemplate) Identifier() string {
	return fmt.Sprintf("%s@%s", tmt.TemplateName(), tmt.Version())
}
