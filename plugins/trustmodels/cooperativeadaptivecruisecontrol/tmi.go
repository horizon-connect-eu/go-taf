package cooperativeadaptivecruisecontrol

import (
	"github.com/vs-uulm/go-subjectivelogic/pkg/subjectivelogic"
	"github.com/vs-uulm/go-taf/internal/util"
	"github.com/vs-uulm/go-taf/pkg/core"
	"github.com/vs-uulm/go-taf/pkg/trustmodel/trustmodelupdate"
	"github.com/vs-uulm/taf-tlee-interface/pkg/trustmodelstructure"
)

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

func (e *TrustModelInstance) Template() core.TrustModelTemplate {
	return e.template
}

func (e *TrustModelInstance) Update(update core.Update) {
	//TODO implement me
	switch update := update.(type) {
	case trustmodelupdate.UpdateAtomicTrustOpinion:
		//TODO
		util.UNUSED(update)
	default:
		//ignore
	}
}

func (e *TrustModelInstance) TrustSourceQuantifiers() []core.TrustSourceQuantifier {
	return []core.TrustSourceQuantifier{}
}

func (e *TrustModelInstance) Initialize(params map[string]interface{}) {
	return
}

func (e *TrustModelInstance) Cleanup() {
	return
}

func (e *TrustModelInstance) RTLs() map[string]subjectivelogic.QueryableOpinion {
	return map[string]subjectivelogic.QueryableOpinion{}
}
