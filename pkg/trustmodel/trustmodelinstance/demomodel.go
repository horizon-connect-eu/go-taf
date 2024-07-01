package trustmodelinstance

import (
	"github.com/vs-uulm/go-subjectivelogic/pkg/subjectivelogic"
	internaltrustmodelstructure "github.com/vs-uulm/go-taf/pkg/trustmodel/trustmodelstructure"
	"github.com/vs-uulm/taf-tlee-interface/pkg/trustmodelstructure"
)

type OldExampleTrustModelInstance struct {
	Id          int
	Tmt         string
	Omega1      subjectivelogic.Opinion
	Omega2      subjectivelogic.Opinion
	Version     int
	Fingerprint int
	Omega_DTI_1 subjectivelogic.Opinion
	Omega_DTI_2 subjectivelogic.Opinion
	Weights     map[string]float64
	Evidence1   map[string]bool
	Evidence2   map[string]bool
	RTL1        subjectivelogic.Opinion
	RTL2        subjectivelogic.Opinion
}

func NewTrustModelInstance(id int, tmt string) OldExampleTrustModelInstance {

	dti1, _ := subjectivelogic.NewOpinion(0.2, 0.1, 0.7, 0.5)
	dti2, _ := subjectivelogic.NewOpinion(0.15, 0.15, 0.7, 0.5)

	omega1, _ := subjectivelogic.NewOpinion(0.2, 0.1, 0.7, 0.5)
	omega2, _ := subjectivelogic.NewOpinion(0.15, 0.15, 0.7, 0.5)

	rtl1, _ := subjectivelogic.NewOpinion(0.7, 0.2, 0.1, 0.5)
	rtl2, _ := subjectivelogic.NewOpinion(0.65, 0.25, 0.1, 0.5)

	return OldExampleTrustModelInstance{
		Id:          id,
		Tmt:         tmt,
		Omega_DTI_1: dti1,
		Omega_DTI_2: dti2,
		Weights:     map[string]float64{"SB": 0.15, "IDS": 0.35, "CFI": 0.35},
		Omega1:      omega1,
		Omega2:      omega2,
		Version:     0,
		Fingerprint: -1,
		Evidence1:   make(map[string]bool),
		Evidence2:   make(map[string]bool),
		RTL1:        rtl1,
		RTL2:        rtl2,
	}
}

// structure parameter for runTLEE
func (i *OldExampleTrustModelInstance) GetTrustGraphStructure() trustmodelstructure.TrustGraphStructure {
	return internaltrustmodelstructure.NewTrustGraphDTO("NONE", []trustmodelstructure.AdjacencyListEntry{
		internaltrustmodelstructure.NewAdjacencyEntryDTO("TAF", []string{"ECU1", "ECU2"}),
	})

}

// Values parameter for runTLEE
func (i *OldExampleTrustModelInstance) GetTrustRelationships() map[string][]trustmodelstructure.TrustRelationship {

	return map[string][]trustmodelstructure.TrustRelationship{
		"ECU1": []trustmodelstructure.TrustRelationship{
			internaltrustmodelstructure.NewTrustRelationshipDTO("TAF", "ECU1", &i.Omega1),
		},
		"ECU2": []trustmodelstructure.TrustRelationship{
			internaltrustmodelstructure.NewTrustRelationshipDTO("TAF", "ECU2", &i.Omega2),
		},
	}
}

func (i *OldExampleTrustModelInstance) GetId() int {
	return i.Id
}
