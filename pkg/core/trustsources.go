package core

import "strings"

type EvidenceType int32

const (
	UNKNOWN EvidenceType = iota
	AIV_SECURE_BOOT
	AIV_SECURE_OTA
	AIV_ACCESS_CONTROL
	AIV_APPLICATION_ISOLATION
	AIV_CONTROL_FLOW_INTEGRITY
	MBD_MISBEHAVIOR_REPORT
	TCH_VERIFIABLE_PRESENTATION
)

type TrustSource int32

const (
	NONE TrustSource = iota
	AIV
	MBD
	TCH
)

func (e EvidenceType) String() string {
	switch e {
	case UNKNOWN:
		return "UNKNOWN"
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

func (s TrustSource) String() string {
	switch s {
	case NONE:
		return "NONE"
	case AIV:
		return "AIV"
	case MBD:
		return "MBD"
	case TCH:
		return "TCH"
	default:
		return "UNKOWN_SOURCE"
	}
}

func (e EvidenceType) Source() TrustSource {
	switch e {
	case UNKNOWN:
		return NONE
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

func EvidenceTypeByName(name string) EvidenceType {
	switch {
	case strings.ToUpper(name) == AIV_SECURE_BOOT.String():
		return AIV_SECURE_BOOT
	case strings.ToUpper(name) == AIV_SECURE_OTA.String():
		return AIV_SECURE_OTA
	case strings.ToUpper(name) == AIV_ACCESS_CONTROL.String():
		return AIV_ACCESS_CONTROL
	case strings.ToUpper(name) == AIV_APPLICATION_ISOLATION.String():
		return AIV_APPLICATION_ISOLATION
	case strings.ToUpper(name) == AIV_CONTROL_FLOW_INTEGRITY.String():
		return AIV_CONTROL_FLOW_INTEGRITY
	case strings.ToUpper(name) == MBD_MISBEHAVIOR_REPORT.String():
		return MBD_MISBEHAVIOR_REPORT
	case strings.ToUpper(name) == TCH_VERIFIABLE_PRESENTATION.String():
		return TCH_VERIFIABLE_PRESENTATION
	default:
		return UNKNOWN
	}
}
