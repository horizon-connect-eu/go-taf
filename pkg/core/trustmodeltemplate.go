package core

type TrustModelTemplate interface {
	TemplateName() string
	Version() string
	Spawn(params map[string]string, context TafContext, channels TafChannels) TrustModelInstance //TODO: check which parameters are really needed
	EvidenceTypes() []EvidenceType
	TrustSourceQuantifiers() []TrustSourceQuantifier
	Description() string
}
