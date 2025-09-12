package trustmodel_to_v0_0_1

import (
	"fmt"
	"github.com/horizon-connect-eu/go-taf/pkg/core"
	internaltrustmodelstructure "github.com/horizon-connect-eu/go-taf/pkg/trustmodel/trustmodelstructure"
	"github.com/horizon-connect-eu/go-taf/pkg/trustmodel/trustmodelupdate"
	"github.com/vs-uulm/go-subjectivelogic/pkg/subjectivelogic"
	"github.com/vs-uulm/taf-tlee-interface/pkg/trustmodelstructure"
	"hash/fnv"
)

type TrustModelInstance struct {
	id      string
	version int

	template TrustModelTemplate

	omega subjectivelogic.Opinion

	currentFingerprint uint32
	targetTrustee      string
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
	return
}

func (tmi *TrustModelInstance) Template() core.TrustModelTemplate {
	return tmi.template
}

func (tmi *TrustModelInstance) Structure() trustmodelstructure.TrustGraphStructure {
	return internaltrustmodelstructure.NewTrustGraphDTO(trustmodelstructure.CumulativeFusion, trustmodelstructure.OppositeBeliefDiscount, []trustmodelstructure.AdjacencyListEntry{
		internaltrustmodelstructure.NewAdjacencyEntryDTO("MEC", []string{trusteeIdentifier(tmi.targetTrustee)}),
	})
}

func (tmi *TrustModelInstance) Values() map[string][]trustmodelstructure.TrustRelationship {
	trusteeOpinion, _ := subjectivelogic.NewOpinion(tmi.omega.Belief(), tmi.omega.Disbelief(), tmi.omega.Uncertainty(), tmi.omega.BaseRate())
	return map[string][]trustmodelstructure.TrustRelationship{
		trusteeIdentifier(tmi.targetTrustee): {
			internaltrustmodelstructure.NewTrustRelationshipDTO("MEC", trusteeIdentifier(tmi.targetTrustee), &trusteeOpinion),
		},
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

func (tmi *TrustModelInstance) updateFingerprint() {

	algorithm := fnv.New32a()
	_, err := algorithm.Write([]byte(tmi.targetTrustee))
	if err == nil {
		tmi.currentFingerprint = algorithm.Sum32()
	}
}

/*
vehicleIdentifier is a helper function to turn a plain identifier into an identifier for vehicles used in the structure.
*/
func trusteeIdentifier(id string) string {
	return fmt.Sprintf("vehicle_%s", id)
}

/* -------------- */

func (tmi *TrustModelInstance) Update(update core.Update) bool {
	oldVersion := tmi.Version()
	switch update := update.(type) {
	case trustmodelupdate.UpdateAtomicTrustOpinion:
		if update.Trustee() == trusteeIdentifier(tmi.targetTrustee) {
			tmi.omega.Modify(update.Opinion().Belief(), update.Opinion().Disbelief(), update.Opinion().Uncertainty(), update.Opinion().BaseRate())
			tmi.version++
		}
	default:
		//ignore
	}
	return oldVersion != tmi.Version()

}

func (tmi *TrustModelInstance) RTLs() map[string]subjectivelogic.QueryableOpinion {
	return map[string]subjectivelogic.QueryableOpinion{
		trusteeIdentifier(tmi.targetTrustee): &DefaultRTL,
	}
}
