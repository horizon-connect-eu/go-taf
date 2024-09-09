package core

type TrustModelTemplate interface {
	TemplateName() string
	Version() string
	Spawn(params map[string]string, context TafContext) (TrustModelInstance, error) //TODO: check which parameters are really needed
	EvidenceTypes() []EvidenceType
	TrustSourceQuantifiers() []TrustSourceQuantifier
	Description() string
}
