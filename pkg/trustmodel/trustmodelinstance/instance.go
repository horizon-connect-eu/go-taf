package trustmodelinstance

import (
	"github.com/vs-uulm/go-taf/pkg/core"
	"github.com/vs-uulm/taf-tlee-interface/pkg/trustmodelstructure"
)

type TrustModelInstance interface {
	ID() string
	Version() int
	Fingerprint() uint32
	Structure() trustmodelstructure.TrustGraphStructure
	Values() map[string][]trustmodelstructure.TrustRelationship
	Template() string
	Update() //TODO
	Init(ctx core.TafContext, channels core.TafChannels)
}
