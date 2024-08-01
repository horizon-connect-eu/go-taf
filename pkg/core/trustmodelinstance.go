package core

import (
	"github.com/vs-uulm/go-subjectivelogic/pkg/subjectivelogic"
	"github.com/vs-uulm/taf-tlee-interface/pkg/trustmodelstructure"
)

type TrustModelInstance interface {
	ID() string
	Version() int
	Fingerprint() uint32
	Structure() trustmodelstructure.TrustGraphStructure
	Values() map[string][]trustmodelstructure.TrustRelationship
	Template() TrustModelTemplate
	Update(update Update)
	Initialize(params map[string]interface{})
	Cleanup()
	RTLs() map[string]subjectivelogic.QueryableOpinion
}
