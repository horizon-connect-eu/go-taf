package trustmodelstructure

import "github.com/vs-uulm/go-subjectivelogic/pkg/subjectivelogic"

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
