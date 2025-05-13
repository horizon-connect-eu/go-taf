package intersectionmovementassist

import (
	"github.com/vs-uulm/go-taf/pkg/trustmodel"
	"github.com/vs-uulm/go-taf/plugins/trustmodels/intersectionmovementassist/trustmodel-ima-standalone-v0.0.1"
	trustmodel_ima_standalone_v0_0_2 "github.com/vs-uulm/go-taf/plugins/trustmodels/intersectionmovementassist/trustmodel-ima-standalone-v0.0.2"
	trustmodel_ntm_standalone_v0_0_1 "github.com/vs-uulm/go-taf/plugins/trustmodels/intersectionmovementassist/trustmodel-ntm-standalone_v0.0.1"
)

func init() {
	trustmodel.RegisterTemplate(trustmodel_ima_standalone_v0_0_1.CreateTrustModelTemplate("IMA_STANDALONE", "0.0.1"))
	trustmodel.RegisterTemplate(trustmodel_ima_standalone_v0_0_2.CreateTrustModelTemplate("IMA_STANDALONE", "0.0.2"))
	trustmodel.RegisterTemplate(trustmodel_ntm_standalone_v0_0_1.CreateTrustModelTemplate("NTM_STANDALONE", "0.0.1"))
}
