package trustassessment

import "github.com/vs-uulm/go-taf/pkg/core"

type TrustModelInstanceTable struct {
	tmis map[string][]string //TODO: replace with trie for more efficient lookup/storage later
}

func CreateTrustModelInstanceTable() *TrustModelInstanceTable {
	return &TrustModelInstanceTable{
		tmis: make(map[string][]string),
	}
}

func (t *TrustModelInstanceTable) RegisterTMI(client string, sessionID string, tmtID string, tmiID string) bool {
	id := core.MergeFullTMIIdentifier(client, sessionID, tmtID, tmiID)
	if _, exists := t.tmis[id]; exists {
		return false
	} else {
		t.tmis[id] = []string{client, sessionID, tmtID, tmiID}
		return true
	}
}

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

func (t *TrustModelInstanceTable) ExistsTMI(client string, sessionID string, tmtID string, tmiID string) bool {
	id := core.MergeFullTMIIdentifier(client, sessionID, tmtID, tmiID)
	if _, exists := t.tmis[id]; exists {
		delete(t.tmis, id)
		return true
	} else {
		return false
	}
}

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

func (t *TrustModelInstanceTable) GetAllTMIs() []string {
	i := 0
	keys := make([]string, len(t.tmis))
	for k := range t.tmis {
		keys[i] = k
		i++
	}
	return keys
}
