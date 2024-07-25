package brussels

import (
	"fmt"
	"github.com/vs-uulm/go-subjectivelogic/pkg/subjectivelogic"
	"github.com/vs-uulm/go-taf/pkg/core"
	"github.com/vs-uulm/go-taf/pkg/trustmodel/trustmodelinstance"
	"github.com/vs-uulm/go-taf/pkg/trustmodel/trustmodeltemplate"
	"math/rand/v2"
)

type TrustModelTemplate struct {
	name    string
	version string
}

func CreateTrustModelTemplate(name string, version string) trustmodeltemplate.TrustModelTemplate {
	return TrustModelTemplate{
		name:    name,
		version: version,
	}
}

func (t TrustModelTemplate) Version() string {
	return t.version
}

func (t TrustModelTemplate) TemplateName() string {
	return t.name
}

func (t TrustModelTemplate) Spawn(params map[string]string, context core.TafContext, channels core.TafChannels) trustmodelinstance.TrustModelInstance {

	omegaDTI1, _ := subjectivelogic.NewOpinion(0.2, 0.1, 0.7, 0.5)
	omegaDTI2, _ := subjectivelogic.NewOpinion(0.15, 0.15, 0.7, 0.5)
	omega1, _ := subjectivelogic.NewOpinion(0.2, 0.1, 0.7, 0.5)
	omega2, _ := subjectivelogic.NewOpinion(0.15, 0.15, 0.7, 0.5)
	rtl1, _ := subjectivelogic.NewOpinion(0.7, 0.2, 0.1, 0.5)
	rtl2, _ := subjectivelogic.NewOpinion(0.65, 0.25, 0.1, 0.5)

	return &TrustModelInstance{
		id:           t.TemplateName() + "@" + t.Version() + "-" + fmt.Sprintf("%000000d", rand.IntN(999999)),
		version:      0,
		template:     t,
		omega1:       omega1,
		omega2:       omega2,
		fingerprint:  -1,
		omega_DTI_1:  omegaDTI1,
		omega_DTI_2:  omegaDTI2,
		weights:      map[string]float64{"SB": 0.15, "IDS": 0.35, "CFI": 0.35},
		evidence1:    make(map[string]bool),
		evidence2:    make(map[string]bool),
		rTL1:         rtl1,
		rTL2:         rtl2,
		trustsources: []string{"AIV"},
	}
}
