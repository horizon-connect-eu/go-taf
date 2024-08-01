package core

import "github.com/vs-uulm/go-subjectivelogic/pkg/subjectivelogic"

/*
An AtlResultSet captures the output of a TLEE computation per proposition in three different versions: SL opinions, projected probabilities, and trust decisions.
*/
type AtlResultSet struct {
	tmiID     string
	sessionID string
	version   int
	slResults map[string]subjectivelogic.QueryableOpinion
	ppResults map[string]float64
	tdResults map[string]bool
}

func CreateAtlResultSet(tmiID string, sessionID string, version int, slResults map[string]subjectivelogic.QueryableOpinion, ppResults map[string]float64, tdResults map[string]bool) AtlResultSet {
	return AtlResultSet{
		tmiID: tmiID, sessionID: sessionID, version: version, slResults: slResults, ppResults: ppResults, tdResults: tdResults,
	}
}

func (r AtlResultSet) TmiID() string {
	return r.tmiID
}
func (r AtlResultSet) SessionID() string {
	return r.sessionID
}

func (r AtlResultSet) Version() int {
	return r.version
}

func (r AtlResultSet) ATLs() map[string]subjectivelogic.QueryableOpinion {
	return r.slResults
}

func (r AtlResultSet) ProjectedProbabilities() map[string]float64 {
	return r.ppResults
}

func (r AtlResultSet) TrustDecisions() map[string]bool {
	return r.tdResults
}
