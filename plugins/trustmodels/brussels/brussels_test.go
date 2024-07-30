package brussels

import (
	"github.com/vs-uulm/go-subjectivelogic/pkg/subjectivelogic"
	"github.com/vs-uulm/go-taf/pkg/core"
	"github.com/vs-uulm/go-taf/pkg/trustmodel/trustmodelupdate"
	"testing"
)

func TestLoadJson(t *testing.T) {
	tmt := CreateTrustModelTemplate("test", "0.0.1")
	context := core.TafContext{}
	channels := core.TafChannels{}

	tmi := tmt.Spawn(make(map[string]string), context, channels)

	// -------------- Check Structure() method ---------------------
	structure := tmi.Structure()
	if structure.Operator() != "NONE" {
		t.Error("Wrong operator specified")
	}

	list := structure.AdjacencyList()
	if len(list) != 1 {
		t.Error("Invalid number of elements in Adjacency List")
	} else {
		entry := list[0]

		if entry.SourceNode() != "TAF" {
			t.Error("Incorrect source node")
		}
		target := entry.TargetNodes()
		if len(target) != 2 {
			t.Error("Incorrect number of elements specified as target nodes")
		} else {
			if target[0] != "VC1" {
				t.Error("Incorrect name of target node one")
			}

			if target[1] != "VC2" {
				t.Error("Incorrect name of target node two")
			}
		}
	}

	// -------------- Check Values() method ---------------------
	values := tmi.Values()
	if len(values) != 2 {
		t.Error("Incorrect number of values in values-map")
	} else {
		if tr1, found := values["VC1"]; found {
			if len(tr1) != 1 {
				t.Error("Wrong number of trust relatonships for VC1")
			} else if tr1[0].Source() != "TAF" {
				t.Error("Wrong trustor specified")
			} else if tr1[0].Destination() != "VC1" {
				t.Error("Wrong trustee specified")
			} else {
				opinion := tr1[0].Opinion()
				if opinion.Belief() != 0.2 || opinion.Disbelief() != 0.1 || opinion.Uncertainty() != 0.7 || opinion.BaseRate() != 0.5 {
					t.Error("Wrong trust opinion specified")
				}
			}
		} else {
			t.Error("VC1 not found - Missing trust relationship")
		}

		if tr2, found := values["VC2"]; found {
			if len(tr2) != 1 {
				t.Error("Wrong number of trust relatonships for VC2")
			} else if tr2[0].Source() != "TAF" {
				t.Error("Wrong trustor specified")
			} else if tr2[0].Destination() != "VC2" {
				t.Error("Wrong trustee specified")
			} else {
				opinion := tr2[0].Opinion()
				if opinion.Belief() != 0.15 || opinion.Disbelief() != 0.15 || opinion.Uncertainty() != 0.7 || opinion.BaseRate() != 0.5 {
					t.Error("Wrong trust opinion specified")
				}
			}
		} else {
			t.Error("VC2 not found - Missing trust relationship")
		}
	}

	// -------------- Check update() method ---------------------
	newOpinion, err := subjectivelogic.NewOpinion(0.5, 0.3, 0.2, 0.5)
	if err != nil {
		t.Error("Trust opinion could not be created")
	} else {
		update := trustmodelupdate.UpdateAtomicTrustOpinion{&newOpinion, "VC1"}
		tmi.Update(update)

		values := tmi.Values()
		requestedOpinion := values["VC1"][0]
		if requestedOpinion.Opinion().Belief() != 0.5 || requestedOpinion.Opinion().Disbelief() != 0.3 || requestedOpinion.Opinion().Uncertainty() != 0.2 || requestedOpinion.Opinion().BaseRate() != 0.5 {
			t.Error("Trust opinion not updated correctly")
		}
	}

	newOpinion2, err2 := subjectivelogic.NewOpinion(0.4, 0.2, 0.4, 0.2)
	if err2 != nil {
		t.Error("Trust opinion could not be created")
	} else {
		update := trustmodelupdate.UpdateAtomicTrustOpinion{&newOpinion2, "VC2"}
		tmi.Update(update)

		values := tmi.Values()
		requestedOpinion := values["VC2"][0]
		if requestedOpinion.Opinion().Belief() != 0.4 || requestedOpinion.Opinion().Disbelief() != 0.2 || requestedOpinion.Opinion().Uncertainty() != 0.4 || requestedOpinion.Opinion().BaseRate() != 0.2 {
			t.Error("Trust opinion not updated correctly")
		}
	}

}
