package core

import "github.com/vs-uulm/go-subjectivelogic/pkg/subjectivelogic"

/*
AtlResultSet captures the output of a TLEE computation per proposition in three different versions: SL opinions, projected probabilities, and trust decisions.
*/
type AtlResultSet struct {
	tmiID     string
	version   int
	slResults map[string]subjectivelogic.QueryableOpinion
	ppResults map[string]float64
	tdResults map[string]TrustDecision
}

func CreateAtlResultSet(tmiID string, version int, slResults map[string]subjectivelogic.QueryableOpinion, ppResults map[string]float64, tdResults map[string]TrustDecision) AtlResultSet {
	return AtlResultSet{
		tmiID:     tmiID,
		version:   version,
		slResults: slResults,
		ppResults: ppResults,
		tdResults: tdResults,
	}
}

/*
TmiID return the short TMI ID (as to be used/queried by the client application).
*/
func (r AtlResultSet) TmiID() string {
	return r.tmiID
}

/*
Version returns the Trust Model Instance version the results are based upon.
*/
func (r AtlResultSet) Version() int {
	return r.version
}

/*
ATLs return a map of all propositions and their ATLs.
*/
func (r AtlResultSet) ATLs() map[string]subjectivelogic.QueryableOpinion {
	return r.slResults
}

/*
ProjectedProbabilities return a map of all propositions and their projected probabilities.
*/
func (r AtlResultSet) ProjectedProbabilities() map[string]float64 {
	return r.ppResults
}

/*
ProjectedProbabilities return a map of all propositions and their trust decisions.
*/
func (r AtlResultSet) TrustDecisions() map[string]TrustDecision {
	return r.tdResults
}
