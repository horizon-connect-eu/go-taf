package trustassessment

import "github.com/horizon-connect-eu/go-taf/pkg/core"

/*
The TrustModelInstanceTable is an internal data structure of the TAM to organize currently existing TMIs and query them.
*/
type TrustModelInstanceTable struct {
	//map full TMI ID -> [client, sessionID, tmtID, tmiID]
	tmis map[string][]string //TODO: replace with trie for more efficient lookup/storage later
}

func CreateTrustModelInstanceTable() *TrustModelInstanceTable {
	return &TrustModelInstanceTable{
		tmis: make(map[string][]string),
	}
}

/*
RegisterTMI adds a new TMI.
*/
func (t *TrustModelInstanceTable) RegisterTMI(client string, sessionID string, tmtID string, tmiID string) bool {
	id := core.MergeFullTMIIdentifier(client, sessionID, tmtID, tmiID)
	if _, exists := t.tmis[id]; exists {
		return false
	} else {
		t.tmis[id] = []string{client, sessionID, tmtID, tmiID}
		return true
	}
}

/*
UnregisterTMI removes a TMI.
*/
func (t *TrustModelInstanceTable) UnregisterTMI(client string, sessionID string, tmtID string, tmiID string) bool {
	id := core.MergeFullTMIIdentifier(client, sessionID, tmtID, tmiID)
	exists := t.ExistsTMI(client, sessionID, tmtID, tmiID)
	if exists == true {
		delete(t.tmis, id)
		return true
	} else {
		return false
	}
}

/*
ExistsTMI checks whether a TMI already exists.
*/
func (t *TrustModelInstanceTable) ExistsTMI(client string, sessionID string, tmtID string, tmiID string) bool {
	id := core.MergeFullTMIIdentifier(client, sessionID, tmtID, tmiID)
	if _, exists := t.tmis[id]; exists {
		delete(t.tmis, id)
		return true
	} else {
		return false
	}
}

// QueryTMIs allows to query for existing TMIs based on the full TMI ID and returns matches.
// This functions allows wildcards in the TMI search expression, to include any match.
// Examples:
//
//	//*/*/TMT@0.0.1/* -> any TMI of template TMT@0.0.1
//	//clientA/*/*/* -> any TMI of clientA in all sessisons
//	//*/*/TMT-X@0.0.1/19 -> any TMI of template TMT@0.0.1 and ID 19
//
// ...
func (t *TrustModelInstanceTable) QueryTMIs(query string) ([]string, error) {
	clientQueryPart, sessionQueryPart, tmtQueryPart, tmiQueryPart := core.SplitFullTMIIdentifier(query)
	//TODO: check whether parts are valid,

	results := make([]string, 0)

	//TODO: later use trie to prevent O(N) in all cases
	for id, parts := range t.tmis {
		if parts[0] == clientQueryPart || clientQueryPart == "*" {
			if parts[1] == sessionQueryPart || sessionQueryPart == "*" {
				if parts[2] == tmtQueryPart || tmtQueryPart == "*" {
					if parts[3] == tmiQueryPart || tmiQueryPart == "*" {
						results = append(results, id)
					}
				}
			}
		}
	}
	return results, nil
}

/*
GetAllTMIs returns a list of all existing TMIs by listing their full identifiers.
*/
func (t *TrustModelInstanceTable) GetAllTMIs() []string {
	i := 0
	keys := make([]string, len(t.tmis))
	for k := range t.tmis {
		keys[i] = k
		i++
	}
	return keys
}
