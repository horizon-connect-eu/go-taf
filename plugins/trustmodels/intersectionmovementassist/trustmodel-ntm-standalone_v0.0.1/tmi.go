package trustmodel_ntm_standalone_v0_0_1

import (
	"fmt"
	"github.com/vs-uulm/go-subjectivelogic/pkg/subjectivelogic"
	"github.com/vs-uulm/go-taf/pkg/core"
	internaltrustmodelstructure "github.com/vs-uulm/go-taf/pkg/trustmodel/trustmodelstructure"
	"github.com/vs-uulm/go-taf/pkg/trustmodel/trustmodelupdate"
	"github.com/vs-uulm/taf-tlee-interface/pkg/trustmodelstructure"
	"hash/fnv"
)

type TrustModelInstance struct {
	id       string
	version  int
	template TrustModelTemplate

	omegaTCH subjectivelogic.Opinion
	omegaMBD subjectivelogic.Opinion
	omega    subjectivelogic.Opinion

	targetTrustee      string
	currentFingerprint uint32

	ewmaAlpha float64
}

func (tmi *TrustModelInstance) ID() string {
	return tmi.id
}

func (tmi *TrustModelInstance) Version() int {
	return tmi.version
}

func (tmi *TrustModelInstance) String() string {
	return core.TMIAsString(tmi)
}

func (tmi *TrustModelInstance) Fingerprint() uint32 {
	return tmi.currentFingerprint
}

func (tmi *TrustModelInstance) Cleanup() {
	//nothing to do here (yet)
	return
}

func (tmi *TrustModelInstance) Template() core.TrustModelTemplate {
	return tmi.template
}

func (tmi *TrustModelInstance) RTLs() map[string]subjectivelogic.QueryableOpinion {
	return map[string]subjectivelogic.QueryableOpinion{
		trusteeIdentifier(tmi.targetTrustee): &DefaultRTL,
	}
}

func (tmi *TrustModelInstance) Structure() trustmodelstructure.TrustGraphStructure {
	return internaltrustmodelstructure.NewTrustGraphDTO(trustmodelstructure.CumulativeFusion, trustmodelstructure.OppositeBeliefDiscount, []trustmodelstructure.AdjacencyListEntry{
		internaltrustmodelstructure.NewAdjacencyEntryDTO("MEC", []string{trusteeIdentifier(tmi.targetTrustee)}),
	})
}

func (tmi *TrustModelInstance) updateFingerprint() {

	algorithm := fnv.New32a()
	_, err := algorithm.Write([]byte(tmi.targetTrustee))
	if err == nil {
		tmi.currentFingerprint = algorithm.Sum32()
	}
}

func (tmi *TrustModelInstance) Initialize(params map[string]interface{}) {
	trusteeID, exists := params["trusteeID"]
	if !exists {
		tmi.targetTrustee = tmi.id
	} else {
		tmi.targetTrustee = trusteeID.(string)
	}

	tmi.version = 0
	tmi.currentFingerprint = 0

	tmi.updateFingerprint()
	return

}

/*
vehicleIdentifier is a helper function to turn a plain identifier into an identifier for vehicles used in the structure.
*/
func trusteeIdentifier(id string) string {
	return fmt.Sprintf("vehicle_%s", id)
}

func (tmi *TrustModelInstance) Values() map[string][]trustmodelstructure.TrustRelationship {
	trusteeOpinion, _ := subjectivelogic.CumulativeFusion(&tmi.omegaMBD, &tmi.omegaTCH)
	return map[string][]trustmodelstructure.TrustRelationship{
		trusteeIdentifier(tmi.targetTrustee): []trustmodelstructure.TrustRelationship{
			internaltrustmodelstructure.NewTrustRelationshipDTO("MEC", trusteeIdentifier(tmi.targetTrustee), &trusteeOpinion),
		},
	}
}

func (tmi *TrustModelInstance) Update(update core.Update) bool {

	oldVersion := tmi.Version()
	println(fmt.Sprintf("+%v", update))
	switch update := update.(type) {
	case trustmodelupdate.UpdateAtomicTrustOpinion:

		if update.Trustee() == trusteeIdentifier(tmi.targetTrustee) {
			if update.TrustSource() == core.TCH {
				tmi.omegaTCH.Modify(update.Opinion().Belief(), update.Opinion().Disbelief(), update.Opinion().Uncertainty(), update.Opinion().BaseRate())
				tmi.version++
			} else if update.TrustSource() == core.MBD {

				belief := (1-tmi.ewmaAlpha)*tmi.omegaMBD.Belief() + tmi.ewmaAlpha*update.Opinion().Belief()
				disbelief := (1-tmi.ewmaAlpha)*tmi.omegaMBD.Disbelief() + tmi.ewmaAlpha*update.Opinion().Disbelief()
				newOpinion, _ := subjectivelogic.NewOpinion(belief, disbelief, 1-(belief+disbelief), update.Opinion().BaseRate())

				tmi.omegaMBD.Modify(newOpinion.Belief(), newOpinion.Disbelief(), newOpinion.Uncertainty(), newOpinion.BaseRate())
				tmi.version++
			}
		}
	default:
		//ignore
	}
	return oldVersion != tmi.Version()

}
