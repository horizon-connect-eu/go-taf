package taskoffloading

import (
	"github.com/horizon-connect-eu/go-taf/pkg/trustmodel"
	to_v0_0_1 "github.com/horizon-connect-eu/go-taf/plugins/trustmodels/taskoffloading/trustmodel-to-v0.0.1"
)

func init() {
	trustmodel.RegisterTemplate(to_v0_0_1.CreateTrustModelTemplate("TO", "0.0.1"))
}
