package message

type EvidenceCollectionMessage struct {
	Timestamp    int    `json:"timestamp"`
	TrustModelID int    `json:"trust_model_id"`
	Trustee      string `json:"trustee"` // Normaly also trustor spezified (not necessary in demo version)
	TS_ID        string `json:"ts_id"`
	Evidence     bool   `json:"evidence"`
}
