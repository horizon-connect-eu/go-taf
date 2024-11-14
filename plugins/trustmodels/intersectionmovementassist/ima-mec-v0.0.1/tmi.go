package ima_mec_v0_0_1

import (
	"github.com/vs-uulm/go-subjectivelogic/pkg/subjectivelogic"
	"github.com/vs-uulm/go-taf/pkg/core"
	"github.com/vs-uulm/taf-tlee-interface/pkg/trustmodelstructure"
)

type TrustModelInstance struct {
	id       string
	version  int
	template TrustModelTemplate

	sourceID      string
	sourceOpinion subjectivelogic.QueryableOpinion            // Opinion V_ego -> V_sourceID
	objects       map[string]subjectivelogic.QueryableOpinion // X : Opinion V_ego -> C_sourceID_{X}

	currentStructure   trustmodelstructure.TrustGraphStructure
	currentValues      map[string][]trustmodelstructure.TrustRelationship
	currentFingerprint uint32
	rtls               map[string]subjectivelogic.QueryableOpinion
	staticRTL          subjectivelogic.QueryableOpinion
}

func (e *TrustModelInstance) ID() string {
	return e.id
}

func (e *TrustModelInstance) Version() int {
	return e.version
}

func (e *TrustModelInstance) Fingerprint() uint32 {
	return e.currentFingerprint
}

func (e *TrustModelInstance) Template() core.TrustModelTemplate {
	return e.template
}

func (e *TrustModelInstance) Update(update core.Update) bool {
	return false
}

func (e *TrustModelInstance) incrementVersion() int {
	e.version = e.version + 1
	return e.version
}

func (e *TrustModelInstance) Initialize(params map[string]interface{}) {
}

func (e *TrustModelInstance) Cleanup() {
	//nothing to do here (yet)
	return
}

func (e *TrustModelInstance) Structure() trustmodelstructure.TrustGraphStructure {
	return e.currentStructure
}

func (e *TrustModelInstance) Values() map[string][]trustmodelstructure.TrustRelationship {
	return nil
}

func (e *TrustModelInstance) RTLs() map[string]subjectivelogic.QueryableOpinion {
	return e.rtls
}

func (e *TrustModelInstance) String() string {
	return core.TMIAsString(e)
}
