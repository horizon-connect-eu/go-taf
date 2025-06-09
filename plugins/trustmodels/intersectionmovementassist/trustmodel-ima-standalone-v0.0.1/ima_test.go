package trustmodel_ima_standalone_v0_0_1

import (
	"fmt"
	"github.com/vs-uulm/go-subjectivelogic/pkg/subjectivelogic"
	"github.com/vs-uulm/go-taf/pkg/config"
	"github.com/vs-uulm/go-taf/pkg/core"
	internaltlee "github.com/vs-uulm/go-taf/pkg/tlee"
	"github.com/vs-uulm/go-taf/pkg/trustdecision"
	"github.com/vs-uulm/go-taf/pkg/trustmodel/trustmodelupdate"
	"log/slog"
	"testing"
)

/*
    ┌─────┐
 ┌──┼V_ego┼──────┐
 │  └──┬──┘      │
 │     │         │
 │  ┌──▼─┐       │
 │  │V_27│       │
 │  └─┬──┴───┐   │
 │    │      │   │
 │    │      │   │
┌▼────▼─┐  ┌─▼───▼─┐
│C_27_27│  │C_27_19│
└───────┘  └───────┘
*/

var RTL_C_27_27, _ = subjectivelogic.NewOpinion(.7, .2, .1, .5)
var RTL_C_27_19, _ = subjectivelogic.NewOpinion(.7, .2, .1, .5)

var RTLmap = map[string]subjectivelogic.QueryableOpinion{
	"C_27_27": &RTL_C_27_27,
	"C_27_19": &RTL_C_27_19,
}

func TestTrustworthy(t *testing.T) {
	update1 := map[core.EvidenceType]interface{}{
		core.TCH_SECURE_BOOT:                          1,
		core.TCH_SECURE_OTA:                           1,
		core.TCH_ACCESS_CONTROL:                       1,
		core.TCH_APPLICATION_ISOLATION:                1,
		core.TCH_CONTROL_FLOW_INTEGRITY:               1,
		core.TCH_CONFIGURATION_INTEGRITY_VERIFICATION: 1,
	}
	update2 := map[core.EvidenceType]interface{}{
		core.MBD_MISBEHAVIOR_REPORT: 0,
	}
	RunTMI(t, update1, update2)
}

func TestUntrustworthy(t *testing.T) {
	update1 := map[core.EvidenceType]interface{}{
		core.TCH_SECURE_BOOT:                          0,
		core.TCH_SECURE_OTA:                           1,
		core.TCH_ACCESS_CONTROL:                       1,
		core.TCH_APPLICATION_ISOLATION:                1,
		core.TCH_CONTROL_FLOW_INTEGRITY:               0,
		core.TCH_CONFIGURATION_INTEGRITY_VERIFICATION: 1,
	}
	update2 := map[core.EvidenceType]interface{}{
		core.MBD_MISBEHAVIOR_REPORT: 63,
	}
	RunTMI(t, update1, update2)
}

func RunTMI(t *testing.T, update1 map[core.EvidenceType]interface{}, update2 map[core.EvidenceType]interface{}) {

	tafContext := createTafContext()
	tlee := internaltlee.SpawnNewTLEE(tafContext.Logger, "", false)
	tmt := CreateTrustModelTemplate("IMA_STANDALONE", "0.0.1")

	// Spawn spawner
	tsqs, _, spawner, err := tmt.Spawn(nil, tafContext)
	if err != nil {
		t.Log(err)
		return
	}

	tchTSQ := tsqs[0]
	mbdTSQ := tsqs[1]

	// Initialize TMI
	tmi, err := spawner.OnNewVehicle("27", nil)
	if err != nil {
		t.Log(err)
		return
	}

	// TMI init
	tmi.Initialize(nil)
	t.Log("Trust model after init:")
	t.Log(tmi.String())

	//Add CPM to add C_27_19
	tmi.Update(trustmodelupdate.CreateRefreshCPM("27", []string{"19"}))
	t.Log("Trust model after CPM message")
	t.Log(tmi.String())

	// Quantify TCH evidence and send ATO update to TMI
	tmi.Update(
		trustmodelupdate.CreateAtomicTrustOpinionUpdate(tchTSQ.Quantifier(update1), "V_ego", "V_27", core.TCH),
	)
	t.Log("Trust model after TCH evidence update:")
	t.Log(tmi.String())

	// Quantify MBD evidence and send ATO update to TMI
	tmi.Update(
		trustmodelupdate.CreateAtomicTrustOpinionUpdate(mbdTSQ.Quantifier(update2), "V_ego", "C_27_19", core.MBD),
	)
	t.Log("Trust model after MBD evidence update:")
	t.Log(tmi.String())

	// Run TLEE
	atls, err := tlee.RunTLEE(tmi.ID(), tmi.Version(), tmi.Fingerprint(), tmi.Structure(), tmi.Values())
	if err != nil {
		t.Log(err)
		return
	}

	printATLs(t, atls, RTLmap)

}

func createTafContext() core.TafContext {
	return core.TafContext{
		Configuration: config.Configuration{},
		Logger:        slog.Default(),
		Context:       nil,
		Identifier:    "taf",
		Crypto:        nil,
	}
}

func printATLs(t *testing.T, atls map[string]subjectivelogic.QueryableOpinion, rtls map[string]subjectivelogic.QueryableOpinion) {

	for proposition, opinion := range atls {
		decision := "no decision"
		rtl, exists := RTLmap[proposition]
		if exists {
			switch trustdecision.Decide(opinion, rtl) {
			case core.TRUSTWORTHY:
				decision = "trustworthy"
			default:
				decision = "not trustworthy"
			}
		}
		t.Log("[" + proposition + "]" + "\t" + opinion.String() + " \t Decision: " + decision + " (ATL:" + fmt.Sprintf("%.2f", trustdecision.ProjectProbability(opinion)) + " <=?=> RTL: " + fmt.Sprintf("%.2f", trustdecision.ProjectProbability(rtl)) + ")")
	}
}
