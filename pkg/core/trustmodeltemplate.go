package core

/*
TrustModelTemplate (TMT) defines a template for a trust model from which concrete instances (TrustModelInstance) can be
spawned. A TMT specifies the type of trust sources used for a trust model and how evidence gets quantified.
*/
type TrustModelTemplate interface {

	/*
		TemplateName returns the name of the template.
	*/
	TemplateName() string

	/*
		Version returns the implementation version of this template.
	*/
	Version() string

	/*
		Spawn creates a new TrustModelInstance(s). Depending on the TrustModelTemplateType, there are two
		different ways of spawning instances – directly and/or by returning callable spawn functions.
		The DynamicTrustModelInstanceSpawner provides callback functions that will be called upon certain events that
		in turn trigger the dynamic spawning of new TrustModelInstance(s).
		If successful, Spawn also returns the TrustSourceQuantifier(s) to be used for newly spawned TMIs within
		this session.
		In case of an error, the first three return values are nil an error is return.
	*/
	Spawn(params map[string]string, context TafContext) ([]TrustSourceQuantifier, TrustModelInstance, DynamicTrustModelInstanceSpawner, error)

	/*
		EvidenceTypes returns a list of EvidenceType(s) used by instances of this template.
	*/
	EvidenceTypes() []EvidenceType

	/*
		Description returns a textual description of the template.
	*/
	Description() string

	/*
		TrustModelTemplateType returns the TrustModelTemplateType of this TMT.
	*/
	Type() TrustModelTemplateType

	/*
		Identifier returns an identifying string that includes the name and the version of the template.
	*/
	Identifier() string
}

/*
DynamicTrustModelInstanceSpawner is a listener that provides callback functions that will be called upon certain triggers.
A callback function can then spawn a new TrustModelInstance, if appropriate.
*/
type DynamicTrustModelInstanceSpawner interface {
	/*
		OnNewVehicle is callback function called in case a new vehicle has become known.
	*/
	OnNewVehicle(identifier string, params map[string]string) (TrustModelInstance, error)
}

/*
TrustModelTemplateType defines the type of Trust Model Template relevant for spawning TMIs.
*/
type TrustModelTemplateType uint16

const (
	/*
		A STATIC_TRUST_MODEL is a trust model instance that will be spawned with initialization of a session.
	*/
	STATIC_TRUST_MODEL TrustModelTemplateType = iota
	/*
		A VEHICLE_TRIGGERED_TRUST_MODEL is a trust model instance that can be spawned with the appearance of previously unknown vehicles.
	*/
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
