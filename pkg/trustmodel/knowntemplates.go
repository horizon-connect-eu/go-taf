package trustmodel

import (
	"github.com/vs-uulm/go-taf/pkg/trustmodel/models/examplemodel"
	"github.com/vs-uulm/go-taf/pkg/trustmodel/trustmodeltemplate"
)

var KnownTemplates = map[string]trustmodeltemplate.TrustModelTemplate{
	"EXAMPLE@0.0.1": examplemodel.CreateExampleTrustModelTemplate("EXAMPLE", "0.0.1"),
}
