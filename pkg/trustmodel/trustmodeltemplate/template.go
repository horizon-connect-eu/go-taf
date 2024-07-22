package trustmodeltemplate

import (
	"github.com/vs-uulm/go-taf/pkg/core"
	"github.com/vs-uulm/go-taf/pkg/trustmodel/trustmodelinstance"
)

type TrustModelTemplate interface {
	TemplateName() string
	Version() string
	Spawn(params map[string]string, context core.TafContext) trustmodelinstance.TrustModelInstance
}
