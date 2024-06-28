// This file was generated from JSON Schema using quicktype, do not modify it directly.
// To parse and unparse this JSON data, add this code to your project and do:
//
//    mBDNotify, err := UnmarshalMBDNotify(bytes)
//    bytes, err = mBDNotify.Marshal()
//
//    mBDSubscribeRequest, err := UnmarshalMBDSubscribeRequest(bytes)
//    bytes, err = mBDSubscribeRequest.Marshal()
//
//    mBDSubscribeResponse, err := UnmarshalMBDSubscribeResponse(bytes)
//    bytes, err = mBDSubscribeResponse.Marshal()
//
//    mBDUnsubscribeRequest, err := UnmarshalMBDUnsubscribeRequest(bytes)
//    bytes, err = mBDUnsubscribeRequest.Marshal()
//
//    mBDUnsubscribeResponse, err := UnmarshalMBDUnsubscribeResponse(bytes)
//    bytes, err = mBDUnsubscribeResponse.Marshal()

package mbdmsg

import "encoding/json"

func UnmarshalMBDNotify(data []byte) (MBDNotify, error) {
	var r MBDNotify
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *MBDNotify) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalMBDSubscribeRequest(data []byte) (MBDSubscribeRequest, error) {
	var r MBDSubscribeRequest
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *MBDSubscribeRequest) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalMBDSubscribeResponse(data []byte) (MBDSubscribeResponse, error) {
	var r MBDSubscribeResponse
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *MBDSubscribeResponse) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalMBDUnsubscribeRequest(data []byte) (MBDUnsubscribeRequest, error) {
	var r MBDUnsubscribeRequest
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *MBDUnsubscribeRequest) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalMBDUnsubscribeResponse(data []byte) (MBDUnsubscribeResponse, error) {
	var r MBDUnsubscribeResponse
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *MBDUnsubscribeResponse) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type MBDNotify struct {
	CpmReport      CpmReport `json:"CPM_REPORT"`
	SubscriptionID string    `json:"subscriptionId"`
}

type CpmReport struct {
	Content             Content             `json:"content"`
	GenerationTime      float64             `json:"generationTime"`
	Observationlocation Observationlocation `json:"observationlocation"`
	ReporterArteryID    float64             `json:"reporterArteryId"`
	ReporterPseudoID    float64             `json:"reporterPseudoId"`
	Version             float64             `json:"version"`
}

type Content struct {
	ObservationSet []ObservationSet `json:"observationSet"`
	V2XPduEvidence V2XPduEvidence   `json:"V2XPduEvidence"`
}

type ObservationSet struct {
	Check    float64 `json:"check"`
	TargetID float64 `json:"targetId"`
}

type V2XPduEvidence struct {
	ReferenceTime float64 `json:"referenceTime"`
	SourceID      float64 `json:"sourceId"`
}

type Observationlocation struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type MBDSubscribeRequest struct {
	// The certificate (*base64 string*) issued by the IAM, attesting to the correct execution
	// of the TAF within an enclave.
	AttestationCertificate string `json:"attestationCertificate"`
	Subscribe              bool   `json:"subscribe"`
}

type MBDSubscribeResponse struct {
	Error          *string `json:"error,omitempty"`
	SubscriptionID *string `json:"subscriptionId,omitempty"`
	Success        *string `json:"success,omitempty"`
}

type MBDUnsubscribeRequest struct {
	// The certificate (*base64 string*) issued by the IAM, attesting to the correct execution
	// of the TAF within an enclave.
	AttestationCertificate string `json:"attestationCertificate"`
	SubscriptionID         string `json:"subscriptionId"`
}

type MBDUnsubscribeResponse struct {
	Error   *string `json:"error,omitempty"`
	Success *string `json:"success,omitempty"`
}
