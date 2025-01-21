package vehiclecomputermigration

import (
	"github.com/vs-uulm/go-taf/pkg/trustmodel"
	"github.com/vs-uulm/go-taf/plugins/trustmodels/vehiclecomputermigration/trustmodel-vcm-v0.0.1"
)

func init() {
	trustmodel.RegisterTemplate(trustmodel_vcm_v0_0_1.CreateTrustModelTemplate("VCM", "0.0.1", "The Vehicle Computer (VCM) Trust Model is the trust model relevant for the DENSO use-case and used to evaluate the trustworthiness of two vehicle computers on an ego vehicle."))
}
