package intersectionmovementassist

import "github.com/vs-uulm/go-taf/pkg/trustmodel"

func init() {
	trustmodel.RegisterTemplate(CreateTrustModelTemplate("IMA_STANDALONE", "0.0.1"))
}
