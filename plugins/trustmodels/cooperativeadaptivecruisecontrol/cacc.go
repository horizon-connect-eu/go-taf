package cooperativeadaptivecruisecontrol

import (
	"github.com/vs-uulm/go-taf/pkg/trustmodel"
)

func init() {
	trustmodel.RegisterTemplate(CreateTrustModelTemplate("CACC", "0.0.1"))
}
