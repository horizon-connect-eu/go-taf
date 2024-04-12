package trustmodelinstance

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
		Weights:     map[string]float64{"SB": 0.15, "IDS": 0.35, "CFI": 0.35},
		Omega1:      subjectivelogic.Opinion{Belief: 0.2, Disbelief: 0.1, Uncertainty: 0.7, BaseRate: 0.5},
		Omega2:      subjectivelogic.Opinion{Belief: 0.15, Disbelief: 0.15, Uncertainty: 0.7, BaseRate: 0.5},
		Version:     0,
		Fingerprint: -1,
		Evidence1:   make(map[string]bool),
		Evidence2:   make(map[string]bool),
		RTL1:        subjectivelogic.Opinion{Belief: 0.70, Disbelief: 0.20, Uncertainty: 0.1, BaseRate: 0.5},
		RTL2:        subjectivelogic.Opinion{Belief: 0.65, Disbelief: 0.25, Uncertainty: 0.1, BaseRate: 0.5},
	}
}

// TODO: Implement return hardcoded structure of this trust model instance
func (i TrustModelInstance) GetStructure() trustmodelstructure.Structure {
	var ecu1 = trustmodelstructure.Object{
		ID:        "ECU1",
		Operator:  "NONE",
		Relations: nil,
	}
	var ecu2 = trustmodelstructure.Object{
		ID:        "ECU2",
		Operator:  "NONE",
		Relations: nil,
	}
	var taf = trustmodelstructure.Object{
		ID:       "TAF",
		Operator: "NONE",
		Relations: []trustmodelstructure.Relation{
			{
				ID:     "1139-123",
				Target: "ECU1",
			},
			{
				ID:     "1139-124",
				Target: "ECU2",
			},
		},
	}

	return trustmodelstructure.Structure{
		taf, ecu1, ecu2,
	}
}

// TODO: Implement return of all Trust Opinions (values) of this trust model instance
func (i TrustModelInstance) GetValues() map[string]subjectivelogic.Opinion {

	return map[string]subjectivelogic.Opinion{
		"1139-123": i.Omega1,
		"1139-124": i.Omega2,
	}
}

func (i *TrustModelInstance) GetId() int {
	return i.Id
}
