package smtd

import (
	"github.com/horizon-connect-eu/go-taf/pkg/trustmodel"
	smtd_v0_0_1 "github.com/horizon-connect-eu/go-taf/plugins/trustmodels/smtd/trustmodel-smtd-v0.0.1"
)

func init() {
	trustmodel.RegisterTemplate(smtd_v0_0_1.CreateTrustModelTemplate("SMTD", "0.0.1"))
}
