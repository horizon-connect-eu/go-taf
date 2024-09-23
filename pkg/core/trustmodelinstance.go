package core

import (
	"fmt"
	"github.com/vs-uulm/go-subjectivelogic/pkg/subjectivelogic"
	trustmodelstructure2 "github.com/vs-uulm/go-taf/pkg/trustmodel/trustmodelstructure"
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
	String() string
}

func SplitFullTMIIdentifier(identifier string) (client string, sessionID string, tmtID string, tmiID string) {
	parts := strings.Split(identifier, "/")
	return parts[2], parts[3], parts[4], parts[5]
}

func MergeFullTMIIdentifier(client string, sessionID string, tmtID string, tmiID string) string {
	identifier := fmt.Sprintf("//%s/%s/%s/%s", client, sessionID, tmtID, tmiID)
	return identifier
}

func TMIAsString(tmi TrustModelInstance) string {
	graph := trustmodelstructure2.DumpStructure(tmi.Structure())
	values := trustmodelstructure2.DumpValues(tmi.Values())
	output := fmt.Sprintf("Trust Model Instance\n---------------\nInternal ID:\t%s\nTMT:\t%s\nVersion:\t%d\nFingerprint:\t%d\n", tmi.ID(), tmi.Template().Identifier(), tmi.Version(), tmi.Fingerprint())
	output = output + fmt.Sprintf("%s\n", graph)
	output = output + fmt.Sprintf("%s\n", values)
	if tmi.RTLs() != nil {
		output = output + fmt.Sprintf("%v", tmi.RTLs())
	}
	return output
}
