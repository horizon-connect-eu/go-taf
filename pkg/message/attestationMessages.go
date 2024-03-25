package message

type AttestationMessage struct {
	EvidenceCollectionMessage
	Entities          []string
	AttestationStatus []bool
}
