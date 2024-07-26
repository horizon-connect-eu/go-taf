package brussels

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

	omega1                         subjectivelogic.Opinion
	omega2                         subjectivelogic.Opinion
	fingerprint                    int
	omega_DTI_1                    subjectivelogic.Opinion
	omega_DTI_2                    subjectivelogic.Opinion
	weights                        map[string]float64
	evidence1                      map[string]bool
	evidence2                      map[string]bool
	rTL1                           subjectivelogic.Opinion
	rTL2                           subjectivelogic.Opinion
	trustsources                   []string
	trustSourceQuantifierInstances []core.TrustSourceQuantifierInstance
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

func (e *TrustModelInstance) Init() {
	//TODO implement me

}

func (e *TrustModelInstance) TrustSourceQuantifiers() []core.TrustSourceQuantifierInstance {
	return e.trustSourceQuantifierInstances
}

func (e *TrustModelInstance) Cleanup() {
	//TODO implement me
	//panic("implement me")
}
