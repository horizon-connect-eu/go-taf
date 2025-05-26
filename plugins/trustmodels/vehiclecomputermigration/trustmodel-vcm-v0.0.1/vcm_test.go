package trustmodel_vcm_v0_0_1

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

var RTL_VC1, _ = subjectivelogic.NewOpinion(.7, .2, .1, .5)
var RTL_VC2, _ = subjectivelogic.NewOpinion(.7, .2, .1, .5)

var RTLmap = map[string]subjectivelogic.QueryableOpinion{
	"VC1": &RTL_VC1,
	"VC2": &RTL_VC2,
}

func TestTMI(t *testing.T) {

	tafContext := createTafContext()
	tlee := internaltlee.SpawnNewTLEE(tafContext.Logger, "", false)
	tmt := CreateTrustModelTemplate("VCM", "0.0.1", "Testing")

	// Spawn TMI
	tsqs, tmi, _, err := tmt.Spawn(nil, tafContext)
	if err != nil {
		t.Log(err)
		return
	}
	// Initialize TMI
	tmi.Initialize(map[string]interface{}{})

	t.Log("Trust model after init:")
	t.Log(tmi.String())

	aivVC1TSQ := tsqs[0]
	aivVC2TSQ := tsqs[1]

	// Quantify evidence and send ATO update to TMI
	tmi.Update(
		trustmodelupdate.CreateAtomicTrustOpinionUpdate(aivVC1TSQ.Quantifier(map[core.EvidenceType]interface{}{
			core.AIV_SECURE_BOOT:                          1,
			core.AIV_SECURE_OTA:                           1,
			core.AIV_ACCESS_CONTROL:                       1,
			core.AIV_APPLICATION_ISOLATION:                0,
			core.AIV_CONTROL_FLOW_INTEGRITY:               1,
			core.AIV_CONFIGURATION_INTEGRITY_VERIFICATION: 1,
		}), "TAF", "VC1", core.AIV),
	)

	t.Log("Trust model after AIV evidence update:")
	t.Log(tmi.String())

	// Quantify evidence and send ATO update to TMI
	tmi.Update(
		trustmodelupdate.CreateAtomicTrustOpinionUpdate(aivVC2TSQ.Quantifier(map[core.EvidenceType]interface{}{
			core.AIV_SECURE_BOOT:                          1,
			core.AIV_SECURE_OTA:                           0,
			core.AIV_ACCESS_CONTROL:                       0,
			core.AIV_APPLICATION_ISOLATION:                0,
			core.AIV_CONTROL_FLOW_INTEGRITY:               1,
			core.AIV_CONFIGURATION_INTEGRITY_VERIFICATION: 1,
		}), "TAF", "VC2", core.AIV),
	)

	t.Log("Trust model after AIV evidence update:")
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
			case core.NOT_TRUSTWORTHY:
				decision = "not trustworthy"
			}
		}
		t.Log("[" + proposition + "]" + "\t" + opinion.String() + " \t Decision: " + decision + " (ATL:" + fmt.Sprintf("%.2f", trustdecision.ProjectProbability(opinion)) + " <=?=> RTL: " + fmt.Sprintf("%.2f", trustdecision.ProjectProbability(rtl)) + ")")
	}
}
