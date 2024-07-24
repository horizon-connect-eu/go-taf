package cooperativeadaptivecruisecontrol

import "github.com/vs-uulm/taf-tlee-interface/pkg/trustmodelstructure"

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

func (e *TrustModelInstance) Init() {
	//TODO implement me
	//panic("implement me")
}
func (e *TrustModelInstance) Cleanup() {
	//TODO implement me
	//panic("implement me")
}
