package core

import "github.com/vs-uulm/go-subjectivelogic/pkg/subjectivelogic"

type Quantifier func(values map[EvidenceType]int) subjectivelogic.QueryableOpinion

type TrustSourceQuantifier struct {
	Trustee    string
	Trustor    string
	Scope      string
	Evidence   []EvidenceType
	Source     Source
	Quantifier Quantifier
}
