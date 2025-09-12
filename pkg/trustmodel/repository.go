package trustmodel

import "github.com/horizon-connect-eu/go-taf/pkg/core"

var TemplateRepository = map[string]core.TrustModelTemplate{}

func RegisterTemplate(template core.TrustModelTemplate) {
	TemplateRepository[template.Identifier()] = template
}
