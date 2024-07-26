// This file was generated from JSON Schema using quicktype, do not modify it directly.
// To parse and unparse this JSON data, add this code to your project and do:
//
//    message, err := UnmarshalMessage(bytes)
//    bytes, err = message.Marshal()

package tchmsg

import "encoding/json"

func UnmarshalMessage(data []byte) (Message, error) {
	var r Message
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *Message) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type Message struct {
	Evidence  Evidence  `json:"evidence"`
	TchReport TchReport `json:"tchReport"`
}

type Evidence struct {
	KeyRef                 string `json:"keyRef"`
	Signature              string `json:"signature"`
	SignatureAlgorithmType string `json:"signatureAlgorithmType"`
	Timestamp              string `json:"timestamp"`
}

type TchReport struct {
	AivEvidence    AivEvidence     `json:"aivEvidence"`
	TrusteeReports []TrusteeReport `json:"trusteeReports"`
}

type AivEvidence struct {
	KeyRef                 string `json:"keyRef"`
	Nonce                  string `json:"nonce"`
	Signature              string `json:"signature"`
	SignatureAlgorithmType string `json:"signatureAlgorithmType"`
	Timestamp              string `json:"timestamp"`
}

type TrusteeReport struct {
	AttestationReport []AttestationReport `json:"attestationReport"`
	TrusteeID         string              `json:"trusteeID"`
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
