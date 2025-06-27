// This file was generated from JSON Schema using quicktype, do not modify it directly.
// To parse and unparse this JSON data, add this code to your project and do:
//
//    tchInitRequest, err := UnmarshalTchInitRequest(bytes)
//    bytes, err = tchInitRequest.Marshal()
//
//    tchInitResponse, err := UnmarshalTchInitResponse(bytes)
//    bytes, err = tchInitResponse.Marshal()
//
//    tchNotify, err := UnmarshalTchNotify(bytes)
//    bytes, err = tchNotify.Marshal()
//
//    tchTcRequest, err := UnmarshalTchTcRequest(bytes)
//    bytes, err = tchTcRequest.Marshal()
//
//    tasTcResponse, err := UnmarshalTasTcResponse(bytes)
//    bytes, err = tasTcResponse.Marshal()

package tchmsg

import "encoding/json"

func UnmarshalTchInitRequest(data []byte) (TchInitRequest, error) {
	var r TchInitRequest
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *TchInitRequest) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalTchInitResponse(data []byte) (TchInitResponse, error) {
	var r TchInitResponse
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *TchInitResponse) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalTchNotify(data []byte) (TchNotify, error) {
	var r TchNotify
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *TchNotify) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalTchTcRequest(data []byte) (TchTcRequest, error) {
	var r TchTcRequest
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *TchTcRequest) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalTasTcResponse(data []byte) (TasTcResponse, error) {
	var r TasTcResponse
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *TasTcResponse) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type TchInitRequest struct {
	Evidence *TCHINITREQUESTEvidence `json:"evidence,omitempty"`
	// A non-empty list of pseudonym identifier(s) corresponding to vehicle(s).
	Query []TCHINITREQUESTQuery `json:"query"`
}

type TCHINITREQUESTEvidence struct {
	// Identifier to the public key associated with the MBD component. To be provided by the MBD
	// crypto library.
	KeyRef string `json:"keyRef"`
	// Challenge in HEX format. To be provided by the MBD crypto library.
	Nonce string `json:"nonce"`
	// Signature of the query structure and the rest of the evidence attributes. To be provided
	// by the MBD crypto library.
	Signature string `json:"signature"`
	// Signing algorithm. To be provided by the MBD crypto library.
	SignatureAlgorithmType string `json:"signatureAlgorithmType"`
	// Request creation date in UTC. To be provided by the MBD crypto library.
	Timestamp string `json:"timestamp"`
}

// A query record with trustee identifier and claims.
type TCHINITREQUESTQuery struct {
	// This identifier corresponds to the pseudonym associated with an ego vehicle.
	TrusteeIDs []string `json:"trusteeIDs"`
}

type TchInitResponse struct {
	Error   *string `json:"error,omitempty"`
	Success *string `json:"success,omitempty"`
}

type TchNotify struct {
	Evidence TCHNOTIFYEvidence `json:"evidence"`
	// Unique identifier for this message.
	Tag       *string   `json:"tag,omitempty"`
	TchReport TchReport `json:"tchReport"`
}

type TCHNOTIFYEvidence struct {
	KeyRef                 string `json:"keyRef"`
	Signature              string `json:"signature"`
	SignatureAlgorithmType string `json:"signatureAlgorithmType"`
	Timestamp              string `json:"timestamp"`
}

type TchReport struct {
	// This identifier corresponds to the pseudonym associated with an ego vehicle.
	TrusteeID      string          `json:"trusteeID"`
	TrusteeReports []TrusteeReport `json:"trusteeReports"`
}

type TrusteeReport struct {
	AttestationReport []AttestationReport `json:"attestationReport"`
	// This identifier corresponds to the exact component associated with the reported claims in
	// the attestation report. If this field doesn't exist, then the attributes refer to the
	// entire trustee entity (e.g., ego-vehicle).
	ComponentID *string `json:"componentID,omitempty"`
}

type AttestationReport struct {
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
	Timestamp string `json:"timestamp"`
}

type TchTcRequest struct {
	Evidence *TCHTCREQUESTEvidence `json:"evidence,omitempty"`
	// A non-empty list of trustee devices and associated claims as targets.
	Query []TCHTCREQUESTQuery `json:"query"`
}

type TCHTCREQUESTEvidence struct {
	// Identifier to the public key associated with the MBD component. To be provided by the MBD
	// crypto library.
	KeyRef string `json:"keyRef"`
	// Challenge in HEX format. To be provided by the MBD crypto library.
	Nonce string `json:"nonce"`
	// Signature of the query structure and the rest of the evidence attributes. To be provided
	// by the MBD crypto library.
	Signature string `json:"signature"`
	// Signing algorithm. To be provided by the MBD crypto library.
	SignatureAlgorithmType string `json:"signatureAlgorithmType"`
	// Request creation date in UTC. To be provided by the MBD crypto library.
	Timestamp string `json:"timestamp"`
}

// A query record with trustee identifier and claims.
type TCHTCREQUESTQuery struct {
	// A list of specific claims that are requested. If empty, all available claims should be
	// reported.
	RequestedClaims []RequestedClaim `json:"requestedClaims"`
	// List of kafka topic identifiers where the TCH_NOTIFY message shall be sent.
	TchNotifyDestinationTopics []string `json:"tchNotifyDestinationTopics,omitempty"`
	// This identifier corresponds to the pseudonym associated with an ego vehicle.
	TrusteeID string `json:"trusteeID"`
}

// A list of requested claims along with a debug flag
type RequestedClaim struct {
	// The value to be emulated within the TCH_NOTIFY. If not specified, TCH will fetch the
	// actual appraisal related to this claim
	Debug *int64 `json:"debug,omitempty"`
	// Name of the requested claim
	Name *string `json:"name,omitempty"`
}

type TasTcResponse struct {
	Error   *string `json:"error,omitempty"`
	Success *string `json:"success,omitempty"`
}
