package brussels

import (
	"fmt"
	"github.com/vs-uulm/go-subjectivelogic/pkg/subjectivelogic"
	"github.com/vs-uulm/go-taf/pkg/core"
	"math/rand/v2"
)

func quantifier(values map[core.EvidenceType]int, designTimeTrustOp subjectivelogic.QueryableOpinion, existenceWeights map[core.EvidenceType]float64, outputWeights map[core.EvidenceType]int) subjectivelogic.QueryableOpinion {
	println("Hello World from quantifier")
	sl, _ := subjectivelogic.NewOpinion(0.5, .2, .3, 0.5)
	return &sl
}

var trustSourceQuantifiers = []core.TrustSourceQuantifier{
	core.TrustSourceQuantifier{
		Trustor:  "TAF",
		Trustee:  "VC1",
		Scope:    "VC1",
		Source:   core.AIV,
		Evidence: []core.EvidenceType{core.AIV_SECURE_BOOT, core.AIV_SECURE_OTA, core.AIV_ACCESS_CONTROL, core.AIV_APPLICATION_ISOLATION, core.AIV_CONTROL_FLOW_INTEGRITY},
		Quantifier: func(m map[core.EvidenceType]int) subjectivelogic.QueryableOpinion {
			existenceWeights := map[core.EvidenceType]float64{
				core.AIV_SECURE_BOOT: 0.1,
			}
			outputWeights := map[core.EvidenceType]int{
				core.AIV_SECURE_BOOT: 1,
			}
			designTimeTrustOpinion, _ := subjectivelogic.NewOpinion(0.5, 0.5, 0, 0.5)
			return quantifier(m, &designTimeTrustOpinion, existenceWeights, outputWeights)
		},
	},
	core.TrustSourceQuantifier{
		Trustor:  "TAF",
		Trustee:  "VC2",
		Scope:    "VC2",
		Source:   core.AIV,
		Evidence: []core.EvidenceType{core.AIV_SECURE_BOOT, core.AIV_SECURE_OTA, core.AIV_ACCESS_CONTROL, core.AIV_APPLICATION_ISOLATION, core.AIV_CONTROL_FLOW_INTEGRITY},
		Quantifier: func(m map[core.EvidenceType]int) subjectivelogic.QueryableOpinion {
			existenceWeights := map[core.EvidenceType]float64{
				core.AIV_SECURE_BOOT: 0.1,
			}
			outputWeights := map[core.EvidenceType]int{
				core.AIV_SECURE_BOOT: 1,
			}
			designTimeTrustOpinion, _ := subjectivelogic.NewOpinion(0.5, 0.5, 0, 0.5)
			return quantifier(m, &designTimeTrustOpinion, existenceWeights, outputWeights)
		},
	},
}

var trustSources []core.EvidenceType

func init() {

	//Extract list of used trust sources from trustSourceQuantifierInstances
	evidenceMap := make(map[core.EvidenceType]bool)
	for _, quantifier := range trustSourceQuantifiers {
		for _, evidence := range quantifier.Evidence {
			evidenceMap[evidence] = true
		}
	}
	trustSources = make([]core.EvidenceType, len(evidenceMap))
	i := 0
	for k := range evidenceMap {
		trustSources[i] = k
		i++
	}
}

type TrustModelTemplate struct {
	name                   string
	version                string
	trustSourceQuantifiers []core.TrustSourceQuantifier
}

func CreateTrustModelTemplate(name string, version string) core.TrustModelTemplate {
	return TrustModelTemplate{
		name:                   name,
		version:                version,
		trustSourceQuantifiers: trustSourceQuantifiers,
	}
}

func (t TrustModelTemplate) EvidenceTypes() []core.EvidenceType {
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
		weights:                        map[string]float64{"SB": 0.15, "IDS": 0.35, "CFI": 0.35},
		evidence1:                      make(map[string]bool),
		evidence2:                      make(map[string]bool),
		rTL1:                           rtl1,
		rTL2:                           rtl2,
		trustsources:                   []string{"AIV"},
		trustSourceQuantifierInstances: t.trustSourceQuantifiers,
	}
}

func (t TrustModelTemplate) TrustSourceQuantifiers() []core.TrustSourceQuantifier {
	return t.trustSourceQuantifiers
}
