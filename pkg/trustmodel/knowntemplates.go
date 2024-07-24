package trustmodel

import "github.com/vs-uulm/go-taf/pkg/trustmodel/trustmodeltemplate"

/*
var KnownTemplates = map[string]trustmodeltemplate.TrustModelTemplate{
	"EXAMPLE@0.0.1": examplemodel.CreateTrustModelTemplate("EXAMPLE", "0.0.1"),
	"IMA@0.0.1":     intersectionmovementassist.CreateTrustModelTemplate("IMA", "0.0.1"),
	"CACC@0.0.1":    cooperativeadaptivecruisecontrol.CreateTrustModelTemplate("CACC", "0.0.1"),
}
*/

var TemplateRepository = map[string]trustmodeltemplate.TrustModelTemplate{}

func RegisterTemplate(template trustmodeltemplate.TrustModelTemplate) {
	TemplateRepository[template.TemplateName()+"@"+template.Version()] = template
}
