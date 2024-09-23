package tlee

import (
	"github.com/vs-uulm/go-subjectivelogic/pkg/subjectivelogic"
	trustmodelstructure2 "github.com/vs-uulm/go-taf/pkg/trustmodel/trustmodelstructure"
	"github.com/vs-uulm/taf-tlee-interface/pkg/trustmodelstructure"
	"log/slog"
)

type TLEE struct {
	Logger *slog.Logger
}

func (t *TLEE) RunTLEE(trustmodelID string, version int, fingerprint uint32, structure trustmodelstructure.TrustGraphStructure, values map[string][]trustmodelstructure.TrustRelationship) map[string]subjectivelogic.QueryableOpinion {

	t.Logger.Info("TLEE Input", "Graph Structure", trustmodelstructure2.DumpStructure(structure))
	results := make(map[string]subjectivelogic.QueryableOpinion)

	for _, list := range values {
		for _, relationship := range list {
			results[relationship.Destination()] = relationship.Opinion()
		}
	}
	return results
}
