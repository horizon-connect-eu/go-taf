package brussels

import (
	"fmt"
	"github.com/vs-uulm/go-subjectivelogic/pkg/subjectivelogic"
	"github.com/vs-uulm/go-taf/pkg/core"
	"math/rand/v2"
)

var trustSourceQuantifierInstances = []core.TrustSourceQuantifierInstance{
	core.TrustSourceQuantifierInstance{
		Trustor:  "TAF",
		Trustee:  "VC1",
		Scope:    "VC1",
		Evidence: []core.Evidence{core.AIV_SECURE_BOOT, core.AIV_SECURE_OTA, core.AIV_ACCESS_CONTROL, core.AIV_APPLICATION_ISOLATION, core.AIV_CONTROL_FLOW_INTEGRITY},
	},
	core.TrustSourceQuantifierInstance{
		Trustor:  "TAF",
		Trustee:  "VC2",
		Scope:    "VC2",
		Evidence: []core.Evidence{core.AIV_SECURE_BOOT, core.AIV_SECURE_OTA, core.AIV_ACCESS_CONTROL, core.AIV_APPLICATION_ISOLATION, core.AIV_CONTROL_FLOW_INTEGRITY},
	},
}

var trustSources []core.Evidence

func init() {

	//Extract list of used trust sources from trustSourceQuantifierInstances
	evidenceMap := make(map[core.Evidence]bool)
	for _, quantifier := range trustSourceQuantifierInstances {
		for _, evidence := range quantifier.Evidence {
			evidenceMap[evidence] = true
		}
	}
	trustSources = make([]core.Evidence, len(evidenceMap))
	i := 0
	for k := range evidenceMap {
		trustSources[i] = k
		i++
	}
}

type TrustModelTemplate struct {
	name                           string
	version                        string
	trustSourceQuantifierInstances []core.TrustSourceQuantifierInstance
}

func CreateTrustModelTemplate(name string, version string) core.TrustModelTemplate {
	return TrustModelTemplate{
		name:                           name,
		version:                        version,
		trustSourceQuantifierInstances: trustSourceQuantifierInstances,
	}
}

func (t TrustModelTemplate) EvidenceSources() []core.Evidence {
	return trustSources
}

func (t TrustModelTemplate) Version() string {
	return t.version
}

func (t TrustModelTemplate) TemplateName() string {
	return t.name
}

func (t TrustModelTemplate) Description() string {
	return "TODO: Add description of trust model"
}

func (t TrustModelTemplate) Spawn(params map[string]string, context core.TafContext, channels core.TafChannels) core.TrustModelInstance {

	omegaDTI1, _ := subjectivelogic.NewOpinion(0.2, 0.1, 0.7, 0.5)
	omegaDTI2, _ := subjectivelogic.NewOpinion(0.15, 0.15, 0.7, 0.5)
	omega1, _ := subjectivelogic.NewOpinion(0.2, 0.1, 0.7, 0.5)
	omega2, _ := subjectivelogic.NewOpinion(0.15, 0.15, 0.7, 0.5)
	rtl1, _ := subjectivelogic.NewOpinion(0.7, 0.2, 0.1, 0.5)
	rtl2, _ := subjectivelogic.NewOpinion(0.65, 0.25, 0.1, 0.5)

	return &TrustModelInstance{
		id:                             t.TemplateName() + "@" + t.Version() + "-" + fmt.Sprintf("%000000d", rand.IntN(999999)),
		version:                        0,
		template:                       t,
		omega1:                         omega1,
		omega2:                         omega2,
		fingerprint:                    0,
		omega_DTI_1:                    omegaDTI1,
		omega_DTI_2:                    omegaDTI2,
		weights:                        map[string]float64{"SB": 0.15, "IDS": 0.35, "CFI": 0.35},
		evidence1:                      make(map[string]bool),
		evidence2:                      make(map[string]bool),
		rTL1:                           rtl1,
		rTL2:                           rtl2,
		trustsources:                   []string{"AIV"},
		trustSourceQuantifierInstances: t.trustSourceQuantifierInstances,
	}
}
