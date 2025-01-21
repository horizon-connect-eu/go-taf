package taskoffloading

import (
	"github.com/vs-uulm/go-taf/pkg/trustmodel"
	to_v0_0_1 "github.com/vs-uulm/go-taf/plugins/trustmodels/taskoffloading/to-v0.0.1"
)

func init() {
	trustmodel.RegisterTemplate(to_v0_0_1.CreateTrustModelTemplate("TO", "0.0.1"))
}
