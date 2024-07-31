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
}

func (e *TrustModelInstance) ID() string {
	return e.id
}

func (e *TrustModelInstance) Version() int {
	return e.version
}

func (e *TrustModelInstance) Fingerprint() uint32 {
	return e.fingerprint
}

func (e *TrustModelInstance) Structure() trustmodelstructure.TrustGraphStructure {
	return internaltrustmodelstructure.NewTrustGraphDTO("NONE", []trustmodelstructure.AdjacencyListEntry{
		internaltrustmodelstructure.NewAdjacencyEntryDTO("TAF", []string{"VC1", "VC2"}),
	})
}

func (e *TrustModelInstance) Values() map[string][]trustmodelstructure.TrustRelationship {
	opinionVC1, _ := subjectivelogic.NewOpinion(e.omega1.Belief(), e.omega1.Disbelief(), e.omega1.Uncertainty(), e.omega1.BaseRate())
	opinionVC2, _ := subjectivelogic.NewOpinion(e.omega2.Belief(), e.omega2.Disbelief(), e.omega2.Uncertainty(), e.omega2.BaseRate())
	return map[string][]trustmodelstructure.TrustRelationship{
		"VC1": []trustmodelstructure.TrustRelationship{
			internaltrustmodelstructure.NewTrustRelationshipDTO("TAF", "VC1", &opinionVC1),
		},
		"VC2": []trustmodelstructure.TrustRelationship{
			internaltrustmodelstructure.NewTrustRelationshipDTO("TAF", "VC2", &opinionVC2),
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
