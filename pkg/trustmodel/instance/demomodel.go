package instance

import (
	"github.com/vs-uulm/taf-tlee-interface/pkg/subjectivelogic"
	"github.com/vs-uulm/taf-tlee-interface/pkg/trustmodelstructure"
)

type TrustModelInstance struct {
	id          int
	tmt         string
	omega1      subjectivelogic.Opinion
	omega2      subjectivelogic.Opinion
	version     int
	fingerprint int
	omega_DTI_1 subjectivelogic.Opinion
	omega_DTI_2 subjectivelogic.Opinion
	weights     [3]float64
}

func NewTrustModelInstance(id int, tmt string) TrustModelInstance {
	return TrustModelInstance{
		id:          id,
		tmt:         tmt,
		omega_DTI_1: subjectivelogic.Opinion{Belief: 0.2, Disbelief: 0.1, Uncertainty: 0.7, BaseRate: 0.5},
		omega_DTI_2: subjectivelogic.Opinion{Belief: 0.15, Disbelief: 0.15, Uncertainty: 0.7, BaseRate: 0.5},
		weights:     [3]float64{0.2, 0.4, 0.4},
		omega1:      subjectivelogic.Opinion{Belief: 0.2, Disbelief: 0.1, Uncertainty: 0.7, BaseRate: 0.5},
		omega2:      subjectivelogic.Opinion{Belief: 0.15, Disbelief: 0.15, Uncertainty: 0.7, BaseRate: 0.5},
		version:     0,
		fingerprint: -1,
	}
}

// TODO: Implement return hardcoded structure of this trust model instance
func (i *TrustModelInstance) getStructure() (trustmodelstructure.Structure) {
	return nil
}

// TODO: Implement return of all Trust Opinions (values) of this trust model instance
func (i *TrustModelInstance) getValues() (map[string]subjectivelogic.Opinion) {
	return nil
}

func (i *TrustModelInstance) GetId() int {
	return i.id
}
