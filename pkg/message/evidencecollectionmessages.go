package message

type EvidenceCollectionMessage struct {
	Timestamp    int    `json:"timestamp"`
	TrustModelID int    `json:"trust_model_id"`
	Trustee      string `json:"trustee"`
	TS_ID        string `json:"ts_id"`
	Evidence     bool   `json:"evidence"`
}
