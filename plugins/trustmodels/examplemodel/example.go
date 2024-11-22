package examplemodel

import (
	"github.com/vs-uulm/go-taf/pkg/trustmodel"
	"github.com/vs-uulm/go-taf/plugins/trustmodels/examplemodel/example-v0.0.1"
)

func init() {
	trustmodel.RegisterTemplate(example_v0_0_1.CreateTrustModelTemplate("EXAMPLE", "0.0.1"))
}
