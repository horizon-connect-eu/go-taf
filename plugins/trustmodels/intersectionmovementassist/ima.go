package intersectionmovementassist

import (
	"github.com/vs-uulm/go-taf/pkg/trustmodel"
	"github.com/vs-uulm/go-taf/plugins/trustmodels/intersectionmovementassist/trustmodel-ima-standalone-v0.0.1"
	trustmodel_ima_standalone_v0_0_2 "github.com/vs-uulm/go-taf/plugins/trustmodels/intersectionmovementassist/trustmodel-ima-standalone-v0.0.2"
)

func init() {
	trustmodel.RegisterTemplate(trustmodel_ima_standalone_v0_0_1.CreateTrustModelTemplate("IMA_STANDALONE", "0.0.1"))
	trustmodel.RegisterTemplate(trustmodel_ima_standalone_v0_0_2.CreateTrustModelTemplate("IMA_STANDALONE", "0.0.2"))
	//trustmodel.RegisterTemplate(ima_mec_v0_0_1.CreateTrustModelTemplate("IMA_MEC", "0.0.1"))
}
