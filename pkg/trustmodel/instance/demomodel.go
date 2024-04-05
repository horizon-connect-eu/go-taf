package instance

import (
	"github.com/vs-uulm/taf-tlee-interface/pkg/subjectivelogic"
	"github.com/vs-uulm/taf-tlee-interface/pkg/trustmodelstructure"
)

type TrustModelInstance struct {
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

func NewTrustModelInstance(id int, tmt string) TrustModelInstance {
	return TrustModelInstance{
		Id:          id,
		Tmt:         tmt,
		Omega_DTI_1: subjectivelogic.Opinion{Belief: 0.2, Disbelief: 0.1, Uncertainty: 0.7, BaseRate: 0.5},
		Omega_DTI_2: subjectivelogic.Opinion{Belief: 0.15, Disbelief: 0.15, Uncertainty: 0.7, BaseRate: 0.5},
		Weights:     map[string]float64{"SB": 0.2, "IDS": 0.4, "CFI": 0.4},
		Omega1:      subjectivelogic.Opinion{Belief: 0.2, Disbelief: 0.1, Uncertainty: 0.7, BaseRate: 0.5},
		Omega2:      subjectivelogic.Opinion{Belief: 0.15, Disbelief: 0.15, Uncertainty: 0.7, BaseRate: 0.5},
		Version:     0,
		Fingerprint: -1,
		Evidence1:   make(map[string]bool),
		Evidence2:   make(map[string]bool),
		RTL1:        subjectivelogic.Opinion{Belief: 0.2, Disbelief: 0.1, Uncertainty: 0.7, BaseRate: 0.5},   // RTL1 needs to be updated with reasonable values
		RTL2:        subjectivelogic.Opinion{Belief: 0.15, Disbelief: 0.15, Uncertainty: 0.7, BaseRate: 0.5}, // RTL2 needs to be updated with reasonable values
	}
}

// TODO: Implement return hardcoded structure of this trust model instance
func (i *TrustModelInstance) getStructure() trustmodelstructure.Structure {
	return nil
}

// TODO: Implement return of all Trust Opinions (values) of this trust model instance
func (i *TrustModelInstance) getValues() map[string]subjectivelogic.Opinion {
	return nil
}

func (i *TrustModelInstance) GetId() int {
	return i.Id
}
