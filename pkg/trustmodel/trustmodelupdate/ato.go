package trustmodelupdate

import (
	"encoding/json"
	"github.com/vs-uulm/go-subjectivelogic/pkg/subjectivelogic"
	"github.com/vs-uulm/go-taf/pkg/core"
)

/*
UpdateAtomicTrustOpinion is an TMI update operation that updates the opinion of a trust relationship in a trust model.
*/
type UpdateAtomicTrustOpinion struct {
	opinion     subjectivelogic.QueryableOpinion
	trustSource core.TrustSource
	trustee     string
	trustor     string
}

func (u UpdateAtomicTrustOpinion) Opinion() subjectivelogic.QueryableOpinion {
	return u.opinion
}

func (u UpdateAtomicTrustOpinion) TrustSource() core.TrustSource {
	return u.trustSource
}

func (u UpdateAtomicTrustOpinion) Trustee() string {
	return u.trustee
}

func (u UpdateAtomicTrustOpinion) Trustor() string {
	return u.trustor
}

func CreateAtomicTrustOpinionUpdate(opinion subjectivelogic.QueryableOpinion, trustor string, trustee string, source core.TrustSource) UpdateAtomicTrustOpinion {
	return UpdateAtomicTrustOpinion{
		opinion:     opinion,
		trustSource: source,
		trustee:     trustee,
		trustor:     trustor,
	}
}

func (u UpdateAtomicTrustOpinion) Type() core.UpdateOp {
	return core.UPDATE_ATO
}

func (u UpdateAtomicTrustOpinion) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Opinion     subjectivelogic.QueryableOpinion `json:"opinion"`
		TrustSource string                           `json:"trustSource"`
		Trustor     string                           `json:"trustor"`
		Trustee     string                           `json:"trustee"`
		Update      string                           `json:"update"`
	}{
		Opinion:     u.Opinion(),
		TrustSource: u.trustSource.String(),
		Trustor:     u.trustor,
		Trustee:     u.trustee,
		Update:      u.Type().String(),
	})
}
