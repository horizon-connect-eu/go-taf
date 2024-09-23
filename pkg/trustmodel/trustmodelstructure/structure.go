package trustmodelstructure

import (
	"fmt"
	"github.com/vs-uulm/taf-tlee-interface/pkg/trustmodelstructure"
	"strings"
)

type TrustGraphDTO struct {
	operator      string
	adjacencyList []trustmodelstructure.AdjacencyListEntry
}

func NewTrustGraphDTO(operator string, entries []trustmodelstructure.AdjacencyListEntry) *TrustGraphDTO {
	return &TrustGraphDTO{
		operator:      operator,
		adjacencyList: entries,
	}
}

func (t *TrustGraphDTO) Operator() string {
	return t.operator
}

func (t *TrustGraphDTO) AdjacencyList() []trustmodelstructure.AdjacencyListEntry {
	return t.adjacencyList
}

type AdjacencyEntryDTO struct {
	sourceNode  string
	targetNodes []string
}

func NewAdjacencyEntryDTO(sourceNode string, targetNodes []string) *AdjacencyEntryDTO {
	return &AdjacencyEntryDTO{
		sourceNode:  sourceNode,
		targetNodes: targetNodes,
	}
}

func (a *AdjacencyEntryDTO) SourceNode() string {
	return a.sourceNode
}

func (a *AdjacencyEntryDTO) TargetNodes() []string {
	return a.targetNodes
}

func DumpStructure(structure trustmodelstructure.TrustGraphStructure) string {
	result := []string{"++ Trust Graph Structure ++"}
	result = append(result, "Operator: "+structure.Operator())
	for _, list := range structure.AdjacencyList() {
		result = append(result, list.SourceNode()+"==>"+fmt.Sprintf("%+v", list.TargetNodes()))
	}
	return strings.Join(result, "\n")
}
