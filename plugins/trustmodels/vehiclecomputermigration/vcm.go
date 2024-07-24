package vehiclecomputermigration

import "github.com/vs-uulm/go-taf/pkg/trustmodel"

func init() {
	trustmodel.RegisterTemplate(CreateTrustModelTemplate("VCM", "0.0.1"))
}
