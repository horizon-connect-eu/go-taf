package core

import "github.com/vs-uulm/go-subjectivelogic/pkg/subjectivelogic"

/*
A Quantifier function takes a list of EvidenceType(s) and their concrete evidence values and calculates a single trust opinion.
*/
type Quantifier func(values map[EvidenceType]interface{}) subjectivelogic.QueryableOpinion

/*
A TrustSourceQuantifier specifies how a trust relationship between a trustor and trustee in a specified scope should be
assigned to a matching trust opinion. Therefore, it provides a  Quantifier function that turns EvidenceType(s) into
a trust opinion.
*/
type TrustSourceQuantifier struct {
	Scope       string
	Trustor     string
	Trustee     string
	Evidence    []EvidenceType
	TrustSource TrustSource
	Quantifier  Quantifier
}
