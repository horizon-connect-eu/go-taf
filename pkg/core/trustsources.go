package core

type Evidence int32

const (
	AIV_SECURE_BOOT Evidence = iota
	AIV_SECURE_OTA
	AIV_ACCESS_CONTROL
	AIV_APPLICATION_ISOLATION
	AIV_CONTROL_FLOW_INTEGRITY
	MBD_MISBEHAVIOR_REPORT
	TCH_VERIFIABLE_PRESENTATION
)

type Source int32

const (
	NONE = iota
	AIV
	MBD
	TCH
)

func (e Evidence) String() string {
	switch e {
	case AIV_SECURE_BOOT:
		return "SECURE_BOOT"
	case AIV_SECURE_OTA:
		return "SECURE_OTA"
	case AIV_ACCESS_CONTROL:
		return "ACCESS_CONTROL"
	case AIV_APPLICATION_ISOLATION:
		return "APPLICATION_ISOLATION"
	case AIV_CONTROL_FLOW_INTEGRITY:
		return "CONTROL_FLOW_INTEGRITY"
	case MBD_MISBEHAVIOR_REPORT:
		return "MISBEHAVIOR_REPORT"
	case TCH_VERIFIABLE_PRESENTATION:
		return "VERIFIABLE_PRESENTATION"
	default:
		return "UNKOWN_EVIDENCE"
	}
}

func (e Evidence) Source() Source {
	switch e {
	case AIV_SECURE_BOOT:
		return AIV
	case AIV_SECURE_OTA:
		return AIV
	case AIV_ACCESS_CONTROL:
		return AIV
	case AIV_APPLICATION_ISOLATION:
		return AIV
	case AIV_CONTROL_FLOW_INTEGRITY:
		return AIV
	case MBD_MISBEHAVIOR_REPORT:
		return MBD
	case TCH_VERIFIABLE_PRESENTATION:
		return TCH
	default:
		return NONE
	}
}
