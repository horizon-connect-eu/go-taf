package trustmodelstructure

import "github.com/vs-uulm/taf-tlee-interface/pkg/trustmodelstructure"

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
