package intersectionmovementassist

import (
	"github.com/vs-uulm/go-taf/pkg/trustmodel"
	"github.com/vs-uulm/go-taf/plugins/trustmodels/intersectionmovementassist/ima-standalone-v0.0.1"
)

func init() {
	trustmodel.RegisterTemplate(ima_standalone_v0_0_1.CreateTrustModelTemplate("IMA_STANDALONE", "0.0.1"))
	//trustmodel.RegisterTemplate(ima_mec_v0_0_1.CreateTrustModelTemplate("IMA_MEC", "0.0.1"))
}
