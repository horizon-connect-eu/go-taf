package examplemodel

import (
	"github.com/horizon-connect-eu/go-taf/pkg/trustmodel"
	"github.com/horizon-connect-eu/go-taf/plugins/trustmodels/examplemodel/trustmodel-example-v0.0.1"
)

func init() {
	trustmodel.RegisterTemplate(trustmodel_example_v0_0_1.CreateTrustModelTemplate("EXAMPLE", "0.0.1"))
}
