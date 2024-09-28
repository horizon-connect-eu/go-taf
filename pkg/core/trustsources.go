package core

import "strings"

type EvidenceType uint16

const (
	UNKNOWN EvidenceType = iota
	AIV_SECURE_BOOT
	AIV_SECURE_OTA
	AIV_ACCESS_CONTROL
	AIV_APPLICATION_ISOLATION
	AIV_CONTROL_FLOW_INTEGRITY
	AIV_CONFIGURATION_INTEGRITY_VERIFICATION
	MBD_MISBEHAVIOR_REPORT
	TCH_SECURE_BOOT
	TCH_SECURE_OTA
	TCH_ACCESS_CONTROL
	TCH_APPLICATION_ISOLATION
	TCH_CONTROL_FLOW_INTEGRITY
	TCH_CONFIGURATION_INTEGRITY_VERIFICATION
)

type TrustSource uint16

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
	case AIV_CONFIGURATION_INTEGRITY_VERIFICATION:
		return "CONFIGURATION_INTEGRITY_VERIFICATION"
	case MBD_MISBEHAVIOR_REPORT:
		return "MISBEHAVIOR_REPORT"
	case TCH_SECURE_BOOT:
		return "TCH_SECURE_BOOT"
	case TCH_SECURE_OTA:
		return "TCH_SECURE_OTA"
	case TCH_ACCESS_CONTROL:
		return "TCH_ACCESS_CONTROL"
	case TCH_APPLICATION_ISOLATION:
		return "TCH_APPLICATION_ISOLATION"
	case TCH_CONTROL_FLOW_INTEGRITY:
		return "TCH_CONTROL_FLOW_INTEGRITY"
	case TCH_CONFIGURATION_INTEGRITY_VERIFICATION:
		return "TCH_CONFIGURATION_INTEGRITY_VERIFICATION"
	default:
		return "UNKNOWN_EVIDENCE"
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
		return "UNKNOWN_SOURCE"
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
	case AIV_CONFIGURATION_INTEGRITY_VERIFICATION:
		return AIV
	case MBD_MISBEHAVIOR_REPORT:
		return MBD
	case TCH_SECURE_BOOT:
		return TCH
	case TCH_SECURE_OTA:
		return TCH
	case TCH_ACCESS_CONTROL:
		return TCH
	case TCH_APPLICATION_ISOLATION:
		return TCH
	case TCH_CONTROL_FLOW_INTEGRITY:
		return TCH
	case TCH_CONFIGURATION_INTEGRITY_VERIFICATION:
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
	case strings.ToUpper(name) == AIV_CONFIGURATION_INTEGRITY_VERIFICATION.String():
		return AIV_CONFIGURATION_INTEGRITY_VERIFICATION
	case strings.ToUpper(name) == MBD_MISBEHAVIOR_REPORT.String():
		return MBD_MISBEHAVIOR_REPORT
	case strings.ToUpper(name) == TCH_SECURE_BOOT.String():
		return TCH_SECURE_BOOT
	case strings.ToUpper(name) == TCH_SECURE_OTA.String():
		return TCH_SECURE_OTA
	case strings.ToUpper(name) == TCH_ACCESS_CONTROL.String():
		return TCH_ACCESS_CONTROL
	case strings.ToUpper(name) == TCH_APPLICATION_ISOLATION.String():
		return TCH_APPLICATION_ISOLATION
	case strings.ToUpper(name) == TCH_CONTROL_FLOW_INTEGRITY.String():
		return TCH_CONTROL_FLOW_INTEGRITY
	case strings.ToUpper(name) == TCH_CONFIGURATION_INTEGRITY_VERIFICATION.String():
		return TCH_CONFIGURATION_INTEGRITY_VERIFICATION
	default:
		return UNKNOWN
	}
}
