package core

type TrustModelTemplate interface {
	TemplateName() string
	Version() string
	Spawn(params map[string]string, context TafContext) (TrustModelInstance, DynamicTrustModelInstanceSpawner, error)
	EvidenceTypes() []EvidenceType
	TrustSourceQuantifiers() []TrustSourceQuantifier
	Description() string
	Type() TrustModelTemplateType
	GenerateTrustModelInstanceID(identifiers ...string) string
}

type DynamicTrustModelInstanceSpawner interface {
	OnNewVehicle(identifier string, params map[string]string) (TrustModelInstance, error)
}

type TrustModelTemplateType uint16

const (
	STATIC_TRUST_MODEL TrustModelTemplateType = iota
	VEHICLE_TRIGGERED_TRUST_MODEL
)

func (t TrustModelTemplateType) String() string {
	switch t {
	case STATIC_TRUST_MODEL:
		return "STATIC_TRUST_MODEL"
	case VEHICLE_TRIGGERED_TRUST_MODEL:
		return "VEHICLE_TRIGGERED_TRUST_MODEL"
	default:
		return "UNKNOWN TYPE"
	}
}
