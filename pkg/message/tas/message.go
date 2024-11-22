// This file was generated from JSON Schema using quicktype, do not modify it directly.
// To parse and unparse this JSON data, add this code to your project and do:
//
//    tasInitRequest, err := UnmarshalTasInitRequest(bytes)
//    bytes, err = tasInitRequest.Marshal()
//
//    tasInitResponse, err := UnmarshalTasInitResponse(bytes)
//    bytes, err = tasInitResponse.Marshal()
//
//    tasNotify, err := UnmarshalTasNotify(bytes)
//    bytes, err = tasNotify.Marshal()
//
//    tasSubscribeRequest, err := UnmarshalTasSubscribeRequest(bytes)
//    bytes, err = tasSubscribeRequest.Marshal()
//
//    tasSubscribeResponse, err := UnmarshalTasSubscribeResponse(bytes)
//    bytes, err = tasSubscribeResponse.Marshal()
//
//    tasTaRequest, err := UnmarshalTasTaRequest(bytes)
//    bytes, err = tasTaRequest.Marshal()
//
//    tasTaResponse, err := UnmarshalTasTaResponse(bytes)
//    bytes, err = tasTaResponse.Marshal()
//
//    tasTeardownRequest, err := UnmarshalTasTeardownRequest(bytes)
//    bytes, err = tasTeardownRequest.Marshal()
//
//    tasTeardownResponse, err := UnmarshalTasTeardownResponse(bytes)
//    bytes, err = tasTeardownResponse.Marshal()
//
//    tasTmtDiscover, err := UnmarshalTasTmtDiscover(bytes)
//    bytes, err = tasTmtDiscover.Marshal()
//
//    tasTmtOffer, err := UnmarshalTasTmtOffer(bytes)
//    bytes, err = tasTmtOffer.Marshal()
//
//    tasUnsubscribeRequest, err := UnmarshalTasUnsubscribeRequest(bytes)
//    bytes, err = tasUnsubscribeRequest.Marshal()
//
//    tasUnsubscribeResponse, err := UnmarshalTasUnsubscribeResponse(bytes)
//    bytes, err = tasUnsubscribeResponse.Marshal()

package tasmsg

import "encoding/json"

func UnmarshalTasInitRequest(data []byte) (TasInitRequest, error) {
	var r TasInitRequest
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *TasInitRequest) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalTasInitResponse(data []byte) (TasInitResponse, error) {
	var r TasInitResponse
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *TasInitResponse) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalTasNotify(data []byte) (TasNotify, error) {
	var r TasNotify
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *TasNotify) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalTasSubscribeRequest(data []byte) (TasSubscribeRequest, error) {
	var r TasSubscribeRequest
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *TasSubscribeRequest) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalTasSubscribeResponse(data []byte) (TasSubscribeResponse, error) {
	var r TasSubscribeResponse
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *TasSubscribeResponse) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalTasTaRequest(data []byte) (TasTaRequest, error) {
	var r TasTaRequest
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *TasTaRequest) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalTasTaResponse(data []byte) (TasTaResponse, error) {
	var r TasTaResponse
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *TasTaResponse) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalTasTeardownRequest(data []byte) (TasTeardownRequest, error) {
	var r TasTeardownRequest
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *TasTeardownRequest) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalTasTeardownResponse(data []byte) (TasTeardownResponse, error) {
	var r TasTeardownResponse
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *TasTeardownResponse) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalTasTmtDiscover(data []byte) (TasTmtDiscover, error) {
	var r TasTmtDiscover
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *TasTmtDiscover) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalTasTmtOffer(data []byte) (TasTmtOffer, error) {
	var r TasTmtOffer
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *TasTmtOffer) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalTasUnsubscribeRequest(data []byte) (TasUnsubscribeRequest, error) {
	var r TasUnsubscribeRequest
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *TasUnsubscribeRequest) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalTasUnsubscribeResponse(data []byte) (TasUnsubscribeResponse, error) {
	var r TasUnsubscribeResponse
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *TasUnsubscribeResponse) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type TasInitRequest struct {
	Params             map[string]string `json:"params,omitempty"`
	TrustModelTemplate string            `json:"trustModelTemplate"`
}

type TasInitResponse struct {
	// The certificate (*base64 string*) issued by the IAM, attesting to the correct execution
	// of the TAF within an enclave.
	AttestationCertificate string  `json:"attestationCertificate"`
	Error                  *string `json:"error,omitempty"`
	SessionID              *string `json:"sessionId,omitempty"`
	Success                *string `json:"success,omitempty"`
}

type TasNotify struct {
	// The certificate (*base64 string*) issued by the IAM, attesting to the correct execution
	// of the TAF within an enclave.
	AttestationCertificate string   `json:"attestationCertificate"`
	SessionID              string   `json:"sessionId"`
	SubscriptionID         string   `json:"subscriptionId"`
	Updates                []Update `json:"updates"`
}

type Update struct {
	// The identifier of the trust model instance.
	ID           string              `json:"id"`
	Propositions []UpdateProposition `json:"propositions"`
}

type UpdateProposition struct {
	ActualTrustworthinessLevel []PurpleActualTrustworthinessLevel `json:"actualTrustworthinessLevel"`
	// The identifier of the proposition.
	PropositionID string `json:"propositionId"`
	// The result of the trust decision engine.
	TrustDecision *bool `json:"trustDecision"`
}

type PurpleActualTrustworthinessLevel struct {
	Output PurpleOutput `json:"output"`
	Type   Type         `json:"type"`
}

type PurpleOutput struct {
	BaseRate    *float64 `json:"baseRate,omitempty"`
	Belief      *float64 `json:"belief,omitempty"`
	Disbelief   *float64 `json:"disbelief,omitempty"`
	Uncertainty *float64 `json:"uncertainty,omitempty"`
	Value       *float64 `json:"value,omitempty"`
}

type TasSubscribeRequest struct {
	SessionID string `json:"sessionId"`
	// The query selector to be used for the subscription. If empty, all trust model instances
	// of the session will be used.
	Subscribe Subscribe `json:"subscribe"`
	// The trigger to be used for dispatching notifications upon a change in values.
	Trigger Trigger `json:"trigger"`
}

// The query selector to be used for the subscription. If empty, all trust model instances
// of the session will be used.
type Subscribe struct {
	// A potentially empty list of targets
	Filter []string `json:"filter"`
}

type TasSubscribeResponse struct {
	// The certificate (*base64 string*) issued by the IAM, attesting to the correct execution
	// of the TAF within an enclave.
	AttestationCertificate string  `json:"attestationCertificate"`
	Error                  *string `json:"error,omitempty"`
	SessionID              string  `json:"sessionId"`
	SubscriptionID         *string `json:"subscriptionId,omitempty"`
	Success                *string `json:"success,omitempty"`
}

type TasTaRequest struct {
	// If false, the TAF will recalculate all results without usings its cache.
	AllowCache *bool `json:"allowCache,omitempty"`
	// The query selector
	Query     Query  `json:"query"`
	SessionID string `json:"sessionId"`
}

// The query selector
type Query struct {
	// A potentially empty list of targets
	Filter []string `json:"filter"`
}

type TasTaResponse struct {
	// The certificate (*base64 string*) issued by the IAM, attesting to the correct execution
	// of the TAF within an enclave.
	AttestationCertificate string   `json:"attestationCertificate"`
	Error                  *string  `json:"error,omitempty"`
	Results                []Result `json:"results,omitempty"`
	SessionID              string   `json:"sessionId"`
}

type Result struct {
	// The identifier of the trust model instance.
	ID           string              `json:"id"`
	Propositions []ResultProposition `json:"propositions"`
}

type ResultProposition struct {
	ActualTrustworthinessLevel []FluffyActualTrustworthinessLevel `json:"actualTrustworthinessLevel"`
	// The identifier of the proposition.
	PropositionID string `json:"propositionId"`
	// The result of the trust decision engine.
	TrustDecision *bool `json:"trustDecision"`
}

type FluffyActualTrustworthinessLevel struct {
	Output FluffyOutput `json:"output"`
	Type   Type         `json:"type"`
}

type FluffyOutput struct {
	BaseRate    *float64 `json:"baseRate,omitempty"`
	Belief      *float64 `json:"belief,omitempty"`
	Disbelief   *float64 `json:"disbelief,omitempty"`
	Uncertainty *float64 `json:"uncertainty,omitempty"`
	Value       *float64 `json:"value,omitempty"`
}

type TasTeardownRequest struct {
	SessionID string `json:"sessionId"`
}

type TasTeardownResponse struct {
	// The certificate (*base64 string*) issued by the IAM, attesting to the correct execution
	// of the TAF within an enclave.
	AttestationCertificate string  `json:"attestationCertificate"`
	Error                  *string `json:"error,omitempty"`
	Success                *string `json:"success,omitempty"`
}

type TasTmtDiscover struct {
	TrustModelTemplates bool `json:"trustModelTemplates"`
}

type TasTmtOffer struct {
	TrustModelTemplates map[string]TrustModelTemplate `json:"trustModelTemplates"`
}

type TrustModelTemplate struct {
	Description string `json:"description"`
	Hash        string `json:"hash"`
	Name        string `json:"name"`
	Version     string `json:"version"`
}

type TasUnsubscribeRequest struct {
	SessionID      string `json:"sessionId"`
	SubscriptionID string `json:"subscriptionId"`
}

type TasUnsubscribeResponse struct {
	// The certificate (*base64 string*) issued by the IAM, attesting to the correct execution
	// of the TAF within an enclave.
	AttestationCertificate string  `json:"attestationCertificate"`
	Error                  *string `json:"error,omitempty"`
	SessionID              string  `json:"sessionId"`
	Success                *string `json:"success,omitempty"`
}

type Type string

const (
	ProjectedProbability   Type = "PROJECTED_PROBABILITY"
	SubjectiveLogicOpinion Type = "SUBJECTIVE_LOGIC_OPINION"
)

// The trigger to be used for dispatching notifications upon a change in values.
type Trigger string

const (
	ActualTrustworthinessLevel Trigger = "ACTUAL_TRUSTWORTHINESS_LEVEL"
	TrustDecision              Trigger = "TRUST_DECISION"
)
