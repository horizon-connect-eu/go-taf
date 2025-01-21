package trustmodel_ima_standalone_v0_0_1

import (
	"fmt"
	"github.com/vs-uulm/go-subjectivelogic/pkg/subjectivelogic"
	"github.com/vs-uulm/go-taf/internal/util"
	"github.com/vs-uulm/go-taf/pkg/config"
	"github.com/vs-uulm/go-taf/pkg/core"
	tlee2 "github.com/vs-uulm/go-taf/pkg/tlee"
	internaltrustmodelstructure "github.com/vs-uulm/go-taf/pkg/trustmodel/trustmodelstructure"
	"github.com/vs-uulm/go-taf/pkg/trustmodel/trustmodelupdate"
	"github.com/vs-uulm/taf-tlee-interface/pkg/trustmodelstructure"
	"log/slog"

	"testing"
)

func TestResolve(t *testing.T) {
	t.Log(core.EvidenceTypeBySourceAndName(core.TCH, "SECURE_BOOT"))
}

func TestTMI(t *testing.T) {
	tmt := CreateTrustModelTemplate("IMA_STANDALONE", "0.0.1")
	_, _, spawner, err := tmt.Spawn(nil, core.TafContext{
		Configuration: config.Configuration{},
		Logger:        slog.Default(),
		Context:       nil,
		Identifier:    "taf",
		Crypto:        nil,
	})
	if err != nil {
		t.Log(err)
		return
	}
	tmi, err := spawner.OnNewVehicle("15", nil)
	if err != nil {
		t.Log(err)
		return
	}
	tmi.Initialize(nil)

	updates := []core.Update{
		trustmodelupdate.CreateRefreshCPM("15", []string{"11"}),
		trustmodelupdate.CreateRefreshCPM("15", []string{"17"}),
	}

	t.Log(tmi.String())

	for _, update := range updates {
		tmi.Update(update)
		t.Log(tmi.String())

	}

}

func TestTchTsq(t *testing.T) {
	tsqs, _ := createTrustSourceQuantifiers(map[string]string{})
	tchTSQ := tsqs[0]

	t.Log(tchTSQ.Quantifier(map[core.EvidenceType]int{
		core.TCH_SECURE_BOOT:                          1,
		core.TCH_SECURE_OTA:                           1,
		core.TCH_ACCESS_CONTROL:                       1,
		core.TCH_APPLICATION_ISOLATION:                1,
		core.TCH_CONTROL_FLOW_INTEGRITY:               1,
		core.TCH_CONFIGURATION_INTEGRITY_VERIFICATION: 1,
	}).String())

}

func TestTrustSourceQuantifierFunctions(t *testing.T) {
	param := map[core.EvidenceType]int{
		core.TCH_SECURE_BOOT:                          1,
		core.TCH_ACCESS_CONTROL:                       0,
		core.TCH_CONTROL_FLOW_INTEGRITY:               -1,
		core.TCH_SECURE_OTA:                           0,
		core.TCH_APPLICATION_ISOLATION:                1,
		core.TCH_CONFIGURATION_INTEGRITY_VERIFICATION: -1,
	}

	trustSourceQuantifiers, _ := createTrustSourceQuantifiers(nil)
	sl := trustSourceQuantifiers[0].Quantifier(param)

	t.Logf(sl.String())

	paramMBD := map[core.EvidenceType]int{
		core.MBD_MISBEHAVIOR_REPORT: 3,
	}

	sl2 := trustSourceQuantifiers[1].Quantifier(paramMBD)

	t.Logf(sl2.String())

	params := make(map[string]string)
	params["MBD_ND_SPEE_CONS"] = "2"
	params["MBD_D_SPEE_PLAU"] = "2"
	params["TCH_EXISTENCE_SECURE_BOOT"] = "3"
	params["TCH_OUTPUT_SECURE_BOOT"] = "3"

	trustSourceQuantifiers2, _ := createTrustSourceQuantifiers(params)

	trustSourceQuantifiers2[0].Quantifier(param)
	trustSourceQuantifiers2[1].Quantifier(param)

}

func TestGraph(t *testing.T) {

	fullBelief, _ := subjectivelogic.NewOpinion(1, 0, 0, 0.5)
	fullUncertainty, _ := subjectivelogic.NewOpinion(0, 0, 1, 0.5)
	maybe, _ := subjectivelogic.NewOpinion(.4, .0, .6, 0.5)
	perhaps, _ := subjectivelogic.NewOpinion(.6, .2, .2, 0.5)
	util.UNUSED(fullUncertainty, fullBelief, maybe, perhaps)

	id := "testGraph"
	version := 11
	fingerprint := uint32(4711)

	structure := internaltrustmodelstructure.NewTrustGraphDTO(trustmodelstructure.CumulativeFusion, []trustmodelstructure.AdjacencyListEntry{
		internaltrustmodelstructure.NewAdjacencyEntryDTO("V_ego", []string{"V_19", "C_19_123", "C_19_456", "C_19_19"}),
		internaltrustmodelstructure.NewAdjacencyEntryDTO("V_ego", []string{"C_19_123", "C_19_456", "C_19_19"}),
	})

	values := map[string][]trustmodelstructure.TrustRelationship{
		"C_19_123": []trustmodelstructure.TrustRelationship{
			internaltrustmodelstructure.NewTrustRelationshipDTO("V_ego", "V_19", &maybe),
			internaltrustmodelstructure.NewTrustRelationshipDTO("V_19", "C_19_123", &fullBelief),
			internaltrustmodelstructure.NewTrustRelationshipDTO("V_ego", "C_19_123", &perhaps),
		},
		"C_19_456": []trustmodelstructure.TrustRelationship{
			internaltrustmodelstructure.NewTrustRelationshipDTO("V_ego", "V_19", &perhaps),
			internaltrustmodelstructure.NewTrustRelationshipDTO("V_19", "C_19_456", &fullBelief),
			internaltrustmodelstructure.NewTrustRelationshipDTO("V_ego", "C_19_456", &maybe),
		},
		"C_19_19": []trustmodelstructure.TrustRelationship{
			internaltrustmodelstructure.NewTrustRelationshipDTO("V_ego", "V_19", &perhaps),
			internaltrustmodelstructure.NewTrustRelationshipDTO("V_19", "C_19_19", &fullBelief),
			internaltrustmodelstructure.NewTrustRelationshipDTO("V_ego", "C_19_19", &maybe),
		},
	}

	tlee := &tlee2.TLEE{}
	results, _ := tlee.RunTLEE(id, version, fingerprint, structure, values)

	print(fmt.Sprintf("%v", results))

}
