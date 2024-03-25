package message

type EvidenceCollectionMessage struct {
	TrustModelID  int  `json:"trust_model_id"`
	TrustObjectID int  `json:"trust_object_id"`
	Timestamp     int  `json:"timestamp"`
	Status        bool `json:"status"`
}
