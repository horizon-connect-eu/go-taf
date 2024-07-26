// This file was generated from JSON Schema using quicktype, do not modify it directly.
// To parse and unparse this JSON data, add this code to your project and do:
//
//    aivNotify, err := UnmarshalAivNotify(bytes)
//    bytes, err = aivNotify.Marshal()
//
//    aivRequest, err := UnmarshalAivRequest(bytes)
//    bytes, err = aivRequest.Marshal()
//
//    aivResponse, err := UnmarshalAivResponse(bytes)
//    bytes, err = aivResponse.Marshal()
//
//    aivSubscribeRequest, err := UnmarshalAivSubscribeRequest(bytes)
//    bytes, err = aivSubscribeRequest.Marshal()
//
//    aivSubscribeResponse, err := UnmarshalAivSubscribeResponse(bytes)
//    bytes, err = aivSubscribeResponse.Marshal()
//
//    aivUnsubscribeRequest, err := UnmarshalAivUnsubscribeRequest(bytes)
//    bytes, err = aivUnsubscribeRequest.Marshal()
//
//    aivUnsubscribeResponse, err := UnmarshalAivUnsubscribeResponse(bytes)
//    bytes, err = aivUnsubscribeResponse.Marshal()

package aivmsg

import "encoding/json"

func UnmarshalAivNotify(data []byte) (AivNotify, error) {
	var r AivNotify
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *AivNotify) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalAivRequest(data []byte) (AivRequest, error) {
	var r AivRequest
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *AivRequest) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalAivResponse(data []byte) (AivResponse, error) {
	var r AivResponse
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *AivResponse) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalAivSubscribeRequest(data []byte) (AivSubscribeRequest, error) {
	var r AivSubscribeRequest
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *AivSubscribeRequest) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalAivSubscribeResponse(data []byte) (AivSubscribeResponse, error) {
	var r AivSubscribeResponse
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *AivSubscribeResponse) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalAivUnsubscribeRequest(data []byte) (AivUnsubscribeRequest, error) {
	var r AivUnsubscribeRequest
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *AivUnsubscribeRequest) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalAivUnsubscribeResponse(data []byte) (AivUnsubscribeResponse, error) {
	var r AivUnsubscribeResponse
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *AivUnsubscribeResponse) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type AivNotify struct {
	AivEvidence AIVNOTIFYAivEvidence `json:"aivEvidence"`
	// The unique identifier used for linking notifications with a specific subscription.
	SubscriptionID *string                  `json:"subscriptionId,omitempty"`
	TrusteeReports []AIVNOTIFYTrusteeReport `json:"trusteeReports"`
}

type AIVNOTIFYAivEvidence struct {
	KeyRef string `json:"keyRef"`
	// In Hex.
	Nonce                  string `json:"nonce"`
	Signature              string `json:"signature"`
	SignatureAlgorithmType string `json:"signatureAlgorithmType"`
	Timestamp              string `json:"timestamp"`
}

type AIVNOTIFYTrusteeReport struct {
	// A list of specific claims that are requested. If empty, all available claims should be
	// reported.
	AttestationReport []PurpleAttestationReport `json:"attestationReport,omitempty"`
	// The identifier of the trustee from which claims should be reported.
	TrusteeID *string `json:"trusteeID,omitempty"`
}

type PurpleAttestationReport struct {
	// Verification status based on the attestation mechanisms. Possible values:
	// Value 0: The verifier (i.e., AIV) asserts that the attestation process has failed for a
	// specific claim.
	// Value 1: The verifier (i.e., AIV) affirms that a specific claim has been successfully
	// verified.
	// Value -1: The verifier (i.e., AIV) hasn't engaged in an attestation process for this
	// specific claim (e.g., because the attestation process is not supported or the prover ECU
	// is not responding). Note that this could be treated equivalently by TAF as no claim being
	// made.
	// Value -2: The verifier (i.e., AIV) initiated an attestation process but didn't receive
	// the expected evidence from the prover entity (e.g., request timeout, malformed response).
	Appraisal int64  `json:"appraisal"`
	Claim     string `json:"claim"`
	// Date
	Timestamp string `json:"timestamp"`
}

type AivRequest struct {
	// The certificate (*base64 string*) issued by the IAM, attesting to the correct execution
	// of the TAF within an enclave.
	AttestationCertificate string             `json:"attestationCertificate"`
	Evidence               AIVREQUESTEvidence `json:"evidence"`
	// A non-empty list of trustee devices and associated claims as targets.
	Query []Query `json:"query"`
}

type AIVREQUESTEvidence struct {
	// Identifier to the public key associated with the TAF component. To be provided by the TAF
	// crypto library.
	KeyRef string `json:"keyRef"`
	// Challenge in HEX format. To be provided by the TAF crypto library.
	Nonce string `json:"nonce"`
	// Signature of the query structure and the rest of the evidence attributes. To be provided
	// by the TAF crypto library.
	Signature string `json:"signature"`
	// Signing algorithm. To be provided by the TAF crypto library.
	SignatureAlgorithmType string `json:"signatureAlgorithmType"`
	// Request creation date in UTC. To be provided by the TAF crypto library.
	Timestamp string `json:"timestamp"`
}

// A query record with trustee identifier and claims.
type Query struct {
	// A list of specific claims that are requested. If empty, all available claims should be
	// reported.
	RequestedClaims []string `json:"requestedClaims"`
	// The identifier of the trustee from which claims should be reported.
	TrusteeID string `json:"TrusteeID"`
}

type AivResponse struct {
	AivEvidence    AIVRESPONSEAivEvidence     `json:"aivEvidence"`
	TrusteeReports []AIVRESPONSETrusteeReport `json:"trusteeReports"`
}

type AIVRESPONSEAivEvidence struct {
	KeyRef string `json:"keyRef"`
	// In Hex.
	Nonce                  *string `json:"nonce,omitempty"`
	Signature              string  `json:"signature"`
	SignatureAlgorithmType string  `json:"signatureAlgorithmType"`
	Timestamp              string  `json:"timestamp"`
}

type AIVRESPONSETrusteeReport struct {
	// A list of specific claims that are requested. If empty, all available claims should be
	// reported.
	AttestationReport []FluffyAttestationReport `json:"attestationReport,omitempty"`
	// The identifier of the trustee from which claims should be reported.
	TrusteeID *string `json:"trusteeID,omitempty"`
}

type FluffyAttestationReport struct {
	// Verification status based on the attestation mechanisms. Possible values:
	// Value 0: The verifier (i.e., AIV) asserts that the attestation process has failed for a
	// specific claim.
	// Value 1: The verifier (i.e., AIV) affirms that a specific claim has been successfully
	// verified.
	// Value -1: The verifier (i.e., AIV) hasn't engaged in an attestation process for this
	// specific claim (e.g., because the attestation process is not supported or the prover ECU
	// is not responding). Note that this could be treated equivalently by TAF as no claim being
	// made.
	// Value -2: The verifier (i.e., AIV) initiated an attestation process but didn't receive
	// the expected evidence from the prover entity (e.g., request timeout, malformed response).
	Appraisal int64  `json:"appraisal"`
	Claim     string `json:"claim"`
	// Date
	Timestamp string `json:"timestamp"`
}

type AivSubscribeRequest struct {
	// The certificate (*base64 string*) issued by the IAM, attesting to the correct execution
	// of the TAF within an enclave.
	AttestationCertificate string `json:"attestationCertificate"`
	// The time interval (in ms) in which the TAF assumes the AIV to check the claims
	// periodically.
	CheckInterval int64                       `json:"checkInterval"`
	Evidence      AIVSUBSCRIBEREQUESTEvidence `json:"evidence"`
	// A non-empty list of trustee devices and associated claims as subscription targets.
	Subscribe []Subscribe `json:"subscribe"`
}

type AIVSUBSCRIBEREQUESTEvidence struct {
	// Identifier to the public key associated with the TAF component. To be provided by the TAF
	// crypto library.
	KeyRef string `json:"keyRef"`
	// Challenge in HEX format. To be provided by the TAF crypto library.
	Nonce string `json:"nonce"`
	// Signature of the query structure and the rest of the evidence attributes. To be provided
	// by the TAF crypto library.
	Signature string `json:"signature"`
	// Signing algorithm. To be provided by the TAF crypto library.
	SignatureAlgorithmType string `json:"signatureAlgorithmType"`
	// Request creation date in UTC. To be provided by the TAF crypto library.
	Timestamp string `json:"timestamp"`
}

// A query record with trustee identifier and claims.
type Subscribe struct {
	// A list of specific claims that are requested. If empty, all available claims should be
	// reported.
	RequestedClaims []string `json:"requestedClaims"`
	// The identifier of the trustee from which claims should be reported.
	TrusteeID string `json:"TrusteeID"`
}

type AivSubscribeResponse struct {
	Error          *string `json:"error,omitempty"`
	SubscriptionID *string `json:"subscriptionId,omitempty"`
	Success        *string `json:"success,omitempty"`
}

type AivUnsubscribeRequest struct {
	// The certificate (*base64 string*) issued by the IAM, attesting to the correct execution
	// of the TAF within an enclave.
	AttestationCertificate string `json:"attestationCertificate"`
	SubscriptionID         string `json:"subscriptionId"`
}

type AivUnsubscribeResponse struct {
	Error   *string `json:"error,omitempty"`
	Success *string `json:"success,omitempty"`
}
