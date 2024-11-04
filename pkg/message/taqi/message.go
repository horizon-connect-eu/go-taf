// This file was generated from JSON Schema using quicktype, do not modify it directly.
// To parse and unparse this JSON data, add this code to your project and do:
//
//    taqiQuery, err := UnmarshalTaqiQuery(bytes)
//    bytes, err = taqiQuery.Marshal()
//
//    taqiResult, err := UnmarshalTaqiResult(bytes)
//    bytes, err = taqiResult.Marshal()

package taqimsg

import "encoding/json"

func UnmarshalTaqiQuery(data []byte) (TaqiQuery, error) {
	var r TaqiQuery
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *TaqiQuery) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalTaqiResult(data []byte) (TaqiResult, error) {
	var r TaqiResult
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *TaqiResult) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type TaqiQuery struct {
	// The query parameters
	Query Query `json:"query"`
}

// The query parameters
type Query struct {
	// Identifier of the trust model instance.
	Identifier string `json:"identifier"`
	// A potentially empty list of propositions.
	Propositions []string `json:"propositions"`
	// The trust model template to be used at the target turst model instance.
	Template string `json:"template"`
}

type TaqiResult struct {
	Error   *string  `json:"error,omitempty"`
	Results []Result `json:"results,omitempty"`
}

type Result struct {
	// The identifier of the trust model instance.
	ID           string        `json:"id"`
	Propositions []Proposition `json:"propositions"`
}

type Proposition struct {
	ActualTrustworthinessLevel []ActualTrustworthinessLevel `json:"actualTrustworthinessLevel"`
	// The identifier of the proposition.
	PropositionID string `json:"propositionId"`
	// The result of the trust decision engine.
	TrustDecision *bool `json:"trustDecision"`
}

type ActualTrustworthinessLevel struct {
	Output Output `json:"output"`
	Type   Type   `json:"type"`
}

type Output struct {
	BaseRate    *float64 `json:"baseRate,omitempty"`
	Belief      *float64 `json:"belief,omitempty"`
	Disbelief   *float64 `json:"disbelief,omitempty"`
	Uncertainty *float64 `json:"uncertainty,omitempty"`
	Value       *float64 `json:"value,omitempty"`
}

type Type string

const (
	ProjectedProbability   Type = "PROJECTED_PROBABILITY"
	SubjectiveLogicOpinion Type = "SUBJECTIVE_LOGIC_OPINION"
)
