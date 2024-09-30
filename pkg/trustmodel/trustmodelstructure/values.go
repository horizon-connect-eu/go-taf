package trustmodelstructure

import (
	"github.com/vs-uulm/go-subjectivelogic/pkg/subjectivelogic"
	"github.com/vs-uulm/taf-tlee-interface/pkg/trustmodelstructure"
	"strings"
)

type TrustRelationshipDTO struct {
	source      string
	destination string
	opinion     subjectivelogic.QueryableOpinion
}

func (r *TrustRelationshipDTO) Source() string {
	return r.source
}

func (r *TrustRelationshipDTO) Destination() string {
	return r.destination
}

func (r *TrustRelationshipDTO) Opinion() subjectivelogic.QueryableOpinion {
	return r.opinion
}

func NewTrustRelationshipDTO(source string, destination string, opinion subjectivelogic.QueryableOpinion) *TrustRelationshipDTO {
	return &TrustRelationshipDTO{
		source:      source,
		destination: destination,
		opinion:     opinion,
	}
}

func DumpValues(values map[string][]trustmodelstructure.TrustRelationship) string {
	result := []string{"++ Values ++"}
	for scope, rels := range values {
		for _, rel := range rels {
			result = append(result, "["+scope+"]"+rel.Source()+"==("+rel.Opinion().String()+")==>"+rel.Destination())
		}
	}
	return strings.Join(result, "\n")
}
