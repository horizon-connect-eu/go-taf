package cooperativeadaptivecruisecontrol

import (
	"fmt"
	"github.com/vs-uulm/go-taf/pkg/core"
	"math/rand/v2"
)

type TrustModelTemplate struct {
	name    string
	version string
}

func CreateTrustModelTemplate(name string, version string) core.TrustModelTemplate {
	return TrustModelTemplate{
		name:    name,
		version: version,
	}
}

func (t TrustModelTemplate) EvidenceSources() []core.Evidence {
	return []core.Evidence{}
}

func (t TrustModelTemplate) Version() string {
	return t.version
}

func (t TrustModelTemplate) TemplateName() string {
	return t.name
}

func (t TrustModelTemplate) Spawn(params map[string]string, context core.TafContext, channels core.TafChannels) core.TrustModelInstance {
	return &TrustModelInstance{
		id:       t.TemplateName() + "@" + t.Version() + "-" + fmt.Sprintf("%000000d", rand.IntN(999999)),
		version:  0,
		template: t,
	}
}
