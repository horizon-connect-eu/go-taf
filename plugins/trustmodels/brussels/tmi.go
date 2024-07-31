package brussels

import (
	"github.com/vs-uulm/go-subjectivelogic/pkg/subjectivelogic"
	"github.com/vs-uulm/go-taf/pkg/core"
	internaltrustmodelstructure "github.com/vs-uulm/go-taf/pkg/trustmodel/trustmodelstructure"
	"github.com/vs-uulm/go-taf/pkg/trustmodel/trustmodelupdate"
	"github.com/vs-uulm/taf-tlee-interface/pkg/trustmodelstructure"
)

type TrustModelInstance struct {
	id      string
	version int

	template TrustModelTemplate

	omega1      subjectivelogic.Opinion
	omega2      subjectivelogic.Opinion
	fingerprint uint32
	rTL1        subjectivelogic.Opinion //TODO: Should this be moved into the template instead because it the same for all instances?
	rTL2        subjectivelogic.Opinion //TODO: Should this be moved into the template instead because it the same for all instances?
}

func (e *TrustModelInstance) ID() string {
	return e.id
}

func (e *TrustModelInstance) Version() int {
	//TODO implement me
	return e.version
}

func (e *TrustModelInstance) Fingerprint() uint32 {
	//TODO implement me
	return e.fingerprint
}

func (e *TrustModelInstance) Structure() trustmodelstructure.TrustGraphStructure {
	return internaltrustmodelstructure.NewTrustGraphDTO("NONE", []trustmodelstructure.AdjacencyListEntry{
		internaltrustmodelstructure.NewAdjacencyEntryDTO("TAF", []string{"VC1", "VC2"}),
	})
}

func (e *TrustModelInstance) Values() map[string][]trustmodelstructure.TrustRelationship {
	//TODO: omega1 and omega2 must be copies and must not be pointers to the actual values
	return map[string][]trustmodelstructure.TrustRelationship{
		"VC1": []trustmodelstructure.TrustRelationship{
			internaltrustmodelstructure.NewTrustRelationshipDTO("TAF", "VC1", &e.omega1),
		},
		"VC2": []trustmodelstructure.TrustRelationship{
			internaltrustmodelstructure.NewTrustRelationshipDTO("TAF", "VC2", &e.omega2),
		},
	}
}

func (e *TrustModelInstance) Template() core.TrustModelTemplate {
	return e.template
}

func (e *TrustModelInstance) Update(update core.Update) {
	switch update := update.(type) {
	case trustmodelupdate.UpdateAtomicTrustOpinion:
		if update.Trustee == "VC1" {
			e.omega1.Modify(update.Opinion.Belief(), update.Opinion.Disbelief(), update.Opinion.Uncertainty(), update.Opinion.BaseRate())
			e.version++
		} else if update.Trustee == "VC2" {
			e.omega2.Modify(update.Opinion.Belief(), update.Opinion.Disbelief(), update.Opinion.Uncertainty(), update.Opinion.BaseRate())
			e.version++
		}
	default:
		//ignore
	}
}

func (e *TrustModelInstance) Initialize(params map[string]interface{}) {
	return
}

func (e *TrustModelInstance) Cleanup() {
	return
}
