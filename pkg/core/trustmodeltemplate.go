package core

type TrustModelTemplate interface {
	TemplateName() string
	Version() string
	Spawn(params map[string]string, context TafContext) (TrustModelInstance, DynamicTrustModelInstanceSpawner, error)
	EvidenceTypes() []EvidenceType
	TrustSourceQuantifiers() []TrustSourceQuantifier
	Description() string
}

type DynamicTrustModelInstanceSpawner interface {
	OnNewVehicle(identifier string, params map[string]string) (TrustModelInstance, error)
}
