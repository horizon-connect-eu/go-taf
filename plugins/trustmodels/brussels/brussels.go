package brussels

import (
	"github.com/horizon-connect-eu/go-taf/pkg/trustmodel"
	brussels001 "github.com/horizon-connect-eu/go-taf/plugins/trustmodels/brussels/v0_0_1"
)

func init() {

	trustmodel.RegisterTemplate(brussels001.CreateTrustModelTemplate("BRUSSELS", "0.0.1", "The BRUSSELS Trust Model is a demo model used to evaluate the trustworthiness of two vehicle computers on an ego vehicle."))
}
