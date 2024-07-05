package examplemodel

import (
	"fmt"
	"github.com/vs-uulm/go-taf/pkg/trustmodel/trustmodelinstance"
	"github.com/vs-uulm/go-taf/pkg/trustmodel/trustmodeltemplate"
	"github.com/vs-uulm/taf-tlee-interface/pkg/trustmodelstructure"
	"math/rand/v2"
)

type ExampleTrustModelTemplate struct {
	name    string
	version string
}

func CreateTrustModelTemplate(name string, version string) trustmodeltemplate.TrustModelTemplate {
	return ExampleTrustModelTemplate{
		name:    name,
		version: version,
	}
}

type ExampleTrustModelInstance struct {
	id      string
	version int

	template ExampleTrustModelTemplate
}

func (t ExampleTrustModelTemplate) Version() string {
	return t.version
}

func (t ExampleTrustModelTemplate) TemplateName() string {
	return t.name
}

func (t ExampleTrustModelTemplate) Spawn(params map[string]string) trustmodelinstance.TrustModelInstance {
	return &ExampleTrustModelInstance{
		id:       t.TemplateName() + "@" + t.Version() + "-" + fmt.Sprintf("%000000d", rand.IntN(999999)),
		version:  0,
		template: t,
	}
}

func (e *ExampleTrustModelInstance) ID() string {
	return e.id
}

func (e *ExampleTrustModelInstance) Version() int {
	//TODO implement me
	panic("implement me")
}

func (e *ExampleTrustModelInstance) Fingerprint() uint32 {
	//TODO implement me
	panic("implement me")
}

func (e *ExampleTrustModelInstance) Structure() trustmodelstructure.TrustGraphStructure {
	//TODO implement me
	panic("implement me")
}

func (e *ExampleTrustModelInstance) Values() map[string][]trustmodelstructure.TrustRelationship {
	//TODO implement me
	panic("implement me")
}

func (e *ExampleTrustModelInstance) Template() string {
	return e.template.TemplateName() + "@" + e.template.Version()
}

func (e *ExampleTrustModelInstance) Update() {
	//TODO implement me
	panic("implement me")
}
