package smtd

import (
	"github.com/vs-uulm/go-taf/pkg/trustmodel"
	smtd_v0_0_1 "github.com/vs-uulm/go-taf/plugins/trustmodels/smtd/trustmodel-smtd-v0.0.1"
)

func init() {
	trustmodel.RegisterTemplate(smtd_v0_0_1.CreateTrustModelTemplate("SMTD", "0.0.1"))
}
