package intersectionmovementassist

import (
	"fmt"
	"github.com/vs-uulm/go-subjectivelogic/pkg/subjectivelogic"
	"github.com/vs-uulm/go-taf/internal/util"
	"github.com/vs-uulm/go-taf/pkg/core"
	"github.com/vs-uulm/go-taf/pkg/trustmodel/trustmodelupdate"
	"github.com/vs-uulm/taf-tlee-interface/pkg/trustmodelstructure"
	"log/slog"
)

type TrustModelInstance struct {
	id       string
	version  int
	template TrustModelTemplate

	sourceID string
}

func (e *TrustModelInstance) ID() string {
	return e.id
}

func (e *TrustModelInstance) Version() int {
	return 0
}

func (e *TrustModelInstance) Fingerprint() uint32 {
	//TODO implement me
	//panic("implement me")
	return 0
}

func (e *TrustModelInstance) Structure() trustmodelstructure.TrustGraphStructure {
	//TODO implement me
	//panic("implement me")
	return nil
}

func (e *TrustModelInstance) Values() map[string][]trustmodelstructure.TrustRelationship {
	//TODO implement me
	return nil
}

func (e *TrustModelInstance) Template() core.TrustModelTemplate {
	return e.template
}

func (e *TrustModelInstance) Update(update core.Update) bool {
	//TODO implement me
	switch update := update.(type) {
	case trustmodelupdate.RefreshCPM:
		//TODO: implement
		slog.Warn("Received Update from Source: " + update.SourceID())
		slog.Warn("Contained IDs: " + fmt.Sprintf("%+v", update.Objects()))
	case trustmodelupdate.UpdateAtomicTrustOpinion:
		//TODO
		util.UNUSED(update)
	default:
		//ignore
	}
	return true
}

func (e *TrustModelInstance) TrustSourceQuantifiers() []core.TrustSourceQuantifier {
	return []core.TrustSourceQuantifier{}
}

func (e *TrustModelInstance) Initialize(params map[string]interface{}) {
	//TODO: get source ID from params
	return
}

func (e *TrustModelInstance) Cleanup() {
	return
}

func (e *TrustModelInstance) RTLs() map[string]subjectivelogic.QueryableOpinion {
	return map[string]subjectivelogic.QueryableOpinion{}
}
