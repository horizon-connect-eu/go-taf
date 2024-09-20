package core

import (
	"fmt"
	"github.com/vs-uulm/go-subjectivelogic/pkg/subjectivelogic"
	"github.com/vs-uulm/taf-tlee-interface/pkg/trustmodelstructure"
	"strings"
)

type TrustModelInstance interface {
	ID() string
	Version() int
	Fingerprint() uint32
	Structure() trustmodelstructure.TrustGraphStructure
	Values() map[string][]trustmodelstructure.TrustRelationship
	Template() TrustModelTemplate
	Update(update Update) bool
	Initialize(params map[string]interface{})
	Cleanup()
	RTLs() map[string]subjectivelogic.QueryableOpinion
}

func SplitFullTMIIdentifier(identifier string) (client string, sessionID string, tmtID string, tmiID string) {
	parts := strings.Split(identifier, "/")
	return parts[2], parts[3], parts[4], parts[5]
}

func MergeFullTMIIdentifier(client string, sessionID string, tmtID string, tmiID string) string {
	identifier := fmt.Sprintf("//%s/%s/%s/%s", client, sessionID, tmtID, tmiID)
	return identifier
}
