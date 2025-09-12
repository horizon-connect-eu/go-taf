package core

import (
	"fmt"
	trustmodelstructure2 "github.com/horizon-connect-eu/go-taf/pkg/trustmodel/trustmodelstructure"
	"github.com/vs-uulm/go-subjectivelogic/pkg/subjectivelogic"
	"github.com/vs-uulm/taf-tlee-interface/pkg/trustmodelstructure"
	"strings"
)

/*
TrustModelInstance (TMI) represent a concrete instance of trust model, spawned from a specific TrustModelTemplate.
A TMI is stateful entity to which update operations can be applied on, and from which properties can be queried
in order to run ATL calculations.
*/
type TrustModelInstance interface {
	/*
		ID returns the (short) ID of the trust model instance. This ID is unique inside a session and for each trust model template.
		Globally unique IDs require full TMI IDs that include clients/sessions/TMTs.
	*/
	ID() string

	/*
		Version returns the version number of the state of this TMI. The version number is a logical clock that gets
		increment for each set of change operations executed on the TMI. This includes both changes to the topology of
		the trust model and changes of the opinions.
	*/
	Version() int

	/*
		Fingerprint returns a numerical fingerprint of the Structure of the TMI. The exact number has no semantic meaning
		and should be treated similar to a hash code. However, if the fingerprint of two versions of a TMI are equal,
		their structure is identical.
	*/
	Fingerprint() uint32

	/*
		Structure returns the topological structure of the TMI in form of an adjacency list between the trust objects.
	*/
	Structure() trustmodelstructure.TrustGraphStructure

	/*
		Values returns the trust values of a TMI. More precisely, in returns for each scope a list of trust
		relationships. The trust relationships each specify the opinions between trustor and trustee in that scope.
	*/
	Values() map[string][]trustmodelstructure.TrustRelationship

	/*
		Template returns the TMT this TMI is based upon.
	*/
	Template() TrustModelTemplate

	/*
		Update executes an update operation on the TMI. The bool return value indicates whether this change could
		require a recalculation of ATLs (so whether the TLEE should be called after the update).
	*/
	Update(update Update) bool

	/*
		Initialize is called once when a TMI has been spawned before it receives the first updates. A map of optional
		(runtime) parameters can be passed to Initialize to configure the TMI in addition to parameters used at spawn time.
	*/
	Initialize(params map[string]interface{})

	/*
		Cleanup is called once before a TMI gets removed from a shard worker and destroyed.
	*/
	Cleanup()

	/*
		RTLs returns the required trust levels for each proposition.
	*/
	RTLs() map[string]subjectivelogic.QueryableOpinion

	/*
		String returns a string representation of the TMI.
	*/
	String() string
}

/*
SplitFullTMIIdentifier takes full TMI identifier and returns its components, namely, client, sessionID, TMT@Version, and (short) tmiID.
*/
func SplitFullTMIIdentifier(identifier string) (client string, sessionID string, tmtID string, tmiID string) {
	parts := strings.Split(identifier, "/")
	return parts[2], parts[3], parts[4], parts[5]
}

/*
MergeFullTMIIdentifier creates a full TMI identifier by merging the individual parts.
*/
func MergeFullTMIIdentifier(client string, sessionID string, tmtID string, tmiID string) string {
	identifier := fmt.Sprintf("//%s/%s/%s/%s", client, sessionID, tmtID, tmiID)
	return identifier
}

/*
TMIAsString is a helper function to take a TMI as an input and returns a string representation of that TMI.
*/
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
