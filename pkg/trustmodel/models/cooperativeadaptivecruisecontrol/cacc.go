package cooperativeadaptivecruisecontrol

import (
	"fmt"
	"github.com/vs-uulm/go-taf/pkg/core"
	"github.com/vs-uulm/go-taf/pkg/trustmodel/trustmodelinstance"
	"github.com/vs-uulm/go-taf/pkg/trustmodel/trustmodeltemplate"
	"github.com/vs-uulm/taf-tlee-interface/pkg/trustmodelstructure"
	"math/rand/v2"
)

type TrustModelTemplate struct {
	name    string
	version string
}

func CreateTrustModelTemplate(name string, version string) trustmodeltemplate.TrustModelTemplate {
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

func (t TrustModelTemplate) Spawn(params map[string]string) trustmodelinstance.TrustModelInstance {
	return &TrustModelInstance{
		id:       t.TemplateName() + "@" + t.Version() + "-" + fmt.Sprintf("%000000d", rand.IntN(999999)),
		version:  0,
		template: t,
	}
}

type TrustModelInstance struct {
	id      string
	version int

	template TrustModelTemplate
}

func (e *TrustModelInstance) ID() string {
	return e.id
}

func (e *TrustModelInstance) Version() int {
	//TODO implement me
	panic("implement me")
}

func (e *TrustModelInstance) Fingerprint() uint32 {
	//TODO implement me
	panic("implement me")
}

func (e *TrustModelInstance) Structure() trustmodelstructure.TrustGraphStructure {
	//TODO implement me
	panic("implement me")
}

func (e *TrustModelInstance) Values() map[string][]trustmodelstructure.TrustRelationship {
	//TODO implement me
	panic("implement me")
}

func (e *TrustModelInstance) Template() string {
	return e.template.TemplateName() + "@" + e.template.Version()
}

func (e *TrustModelInstance) Update() {
	//TODO implement me
	panic("implement me")
}

func (e *TrustModelInstance) Init(ctx core.TafContext, channels core.TafChannels) {
	//TODO implement me
	panic("implement me")
}
