package tlee

import (
	"github.com/vs-uulm/go-subjectivelogic/pkg/subjectivelogic"
	"github.com/vs-uulm/taf-tlee-interface/pkg/trustmodelstructure"
)

type TLEE struct {
}

func (t *TLEE) RunTLEE(trustmodelID string, version int, fingerprint uint32, structure trustmodelstructure.TrustGraphStructure, values map[string][]trustmodelstructure.TrustRelationship) map[string]subjectivelogic.QueryableOpinion {
	results := make(map[string]subjectivelogic.QueryableOpinion)

	for _, list := range values {
		for _, relationship := range list {
			results[relationship.Destination()] = relationship.Opinion()
		}
	}
	return results
}
