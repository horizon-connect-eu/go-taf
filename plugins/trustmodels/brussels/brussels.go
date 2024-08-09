package brussels

import "github.com/vs-uulm/go-taf/pkg/trustmodel"

func init() {
	trustmodel.RegisterTemplate(CreateTrustModelTemplate("BRUSSELS", "0.0.1", "The BRUSSELS Trust Model is a demo model used to evaluate the trustworthiness of two vehicle computers on an ego vehicle."))
}
