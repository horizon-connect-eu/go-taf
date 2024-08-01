package brussels

import (
	actualtlee "connect.informatik.uni-ulm.de/coordination/tlee-implementation/pkg/core"
	"fmt"
	"github.com/vs-uulm/go-taf/internal/util"
	"github.com/vs-uulm/go-taf/pkg/core"
	"github.com/vs-uulm/taf-tlee-interface/pkg/tleeinterface"
	"math"
	"testing"
)

func TestLoadJson(t *testing.T) {
	tmt := CreateTrustModelTemplate("test", "0.0.1", "test")
	context := core.TafContext{}
	channels := core.TafChannels{}

	/*tmi, _ := tmt.Spawn(make(map[string]string), context, channels)

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
		update := trustmodelupdate.CreateAtomicTrustOpinionUpdate(&newOpinion, "VC1", core.AIV)
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
		update := trustmodelupdate.CreateAtomicTrustOpinionUpdate(&newOpinion2, "VC2", core.AIV)
		tmi.Update(update)

		values := tmi.Values()
		requestedOpinion := values["VC2"][0]
		if requestedOpinion.Opinion().Belief() != 0.4 || requestedOpinion.Opinion().Disbelief() != 0.2 || requestedOpinion.Opinion().Uncertainty() != 0.4 || requestedOpinion.Opinion().BaseRate() != 0.2 {
			t.Error("Trust opinion not updated correctly")
		}
	}

	//----------------ATO calculation-------------------
	quantifiers := tmi.Template().TrustSourceQuantifiers()
	evidenceMap := make(map[core.EvidenceType]int)

	// Test Scenario1
	evidenceMap[core.AIV_SECURE_BOOT] = 1
	evidenceMap[core.AIV_ACCESS_CONTROL] = -1
	evidenceMap[core.AIV_CONTROL_FLOW_INTEGRITY] = 0

	slOpinion := quantifiers[0].Quantifier(evidenceMap)
	if (math.Round(slOpinion.Belief()*100)/100) != 0.27 || (math.Round(slOpinion.Disbelief()*100)/100) != 0.45 || (math.Round(slOpinion.Uncertainty()*100)/100) != 0.28 {
		t.Error("Incorrect trust opinion")
	}

	// Test Scenario2
	evidenceMap[core.AIV_SECURE_BOOT] = 0
	evidenceMap[core.AIV_ACCESS_CONTROL] = -1
	evidenceMap[core.AIV_CONTROL_FLOW_INTEGRITY] = 0

	slOpinion = quantifiers[0].Quantifier(evidenceMap)
	if (math.Round(slOpinion.Belief()*100)/100) != 0.0 || (math.Round(slOpinion.Disbelief()*100)/100) != 1.0 || (math.Round(slOpinion.Uncertainty()*100)/100) != 0.0 {
		t.Error("Incorrect trust opinion")
	}*/

	//----------------Dynamic weights-------------------
	values_init := make(map[string]string)
	values_init["VC1_EXISTENCE_SECURE_BOOT"] = "0.1"
	values_init["VC1_EXISTENCE_ACCESS_CONTROL"] = "0.1"
	values_init["VC1_EXISTENCE_SECURE_OTA"] = "0.1"
	values_init["VC1_EXISTENCE_APPLICATION_ISOLATION"] = "0.1"
	values_init["VC1_EXISTENCE_CONTROL_FLOW_INTEGRITY"] = "0.1"
	values_init["VC1_EXISTENCE_CONFIGURATION_INTEGRITY_VERIFICATION"] = "0.1"

	values_init["VC2_EXISTENCE_SECURE_BOOT"] = "0.1"
	values_init["VC2_EXISTENCE_ACCESS_CONTROL"] = "0.1"
	values_init["VC2_EXISTENCE_SECURE_OTA"] = "0.1"
	values_init["VC2_EXISTENCE_APPLICATION_ISOLATION"] = "0.1"
	values_init["VC2_EXISTENCE_CONTROL_FLOW_INTEGRITY"] = "0.1"
	values_init["VC2_EXISTENCE_CONFIGURATION_INTEGRITY_VERIFICATION"] = "0.1"

	values_init["VC1_OUTPUT_SECURE_BOOT"] = "0"
	values_init["VC1_OUTPUT_ACCESS_CONTROL"] = "1"
	values_init["VC1_OUTPUT_SECURE_OTA"] = "1"
	values_init["VC1_OUTPUT_APPLICATION_ISOLATION"] = "2"
	values_init["VC1_OUTPUT_CONTROL_FLOW_INTEGRITY"] = "1"
	values_init["VC1_OUTPUT_CONFIGURATION_INTEGRITY_VERIFICATION"] = "1"

	values_init["VC2_OUTPUT_SECURE_BOOT"] = "1"
	values_init["VC2_OUTPUT_ACCESS_CONTROL"] = "1"
	values_init["VC2_OUTPUT_SECURE_OTA"] = "0"
	values_init["VC2_OUTPUT_APPLICATION_ISOLATION"] = "2"
	values_init["VC2_OUTPUT_CONTROL_FLOW_INTEGRITY"] = "1"
	values_init["VC2_OUTPUT_CONFIGURATION_INTEGRITY_VERIFICATION"] = "0"

	values_init["VC1_DTI_BELIEF"] = "0.0"
	values_init["VC1_DTI_DISBELIEF"] = "0.0"
	values_init["VC1_DTI_UNCERTAINTY"] = "1.0"
	values_init["VC1_DTI_BASERATE"] = "0.5"

	/*values_init["VC2_DTI_BELIEF"] = "0.1"
	values_init["VC2_DTI_DISBELIEF"] = "0.2"
	values_init["VC2_DTI_UNCERTAINTY"] = "0.7"
	values_init["VC2_DTI_BASERATE"] = "0.5"*/

	tmi2, _ := tmt.Spawn(values_init, context, channels)

	// Test Scenario2
	evidenceMap2 := make(map[core.EvidenceType]int)

	evidenceMap2[core.AIV_SECURE_BOOT] = 1
	evidenceMap2[core.AIV_ACCESS_CONTROL] = 1
	evidenceMap2[core.AIV_CONTROL_FLOW_INTEGRITY] = 1

	slOpinion2 := tmi2.Template().TrustSourceQuantifiers()[0].Quantifier(evidenceMap2)
	if (math.Round(slOpinion2.Belief()*100)/100) != 0.3 || (math.Round(slOpinion2.Disbelief()*100)/100) != 1.0 || (math.Round(slOpinion2.Uncertainty()*100)/100) != 0.7 {
		t.Error("Incorrect trust opinion")
	}

	//----------------TLEE execution-------------------
	var tlee tleeinterface.TLEE
	tlee = &actualtlee.TLEE{}
	util.UNUSED(tlee)
	tleeResults := tlee.RunTLEE(tmi2.ID(), tmi2.Version(), tmi2.Fingerprint(), tmi2.Structure(), tmi2.Values())
	fmt.Printf("%+v", tleeResults)
}
