package trustsource

import "strings"

/*
MisbehaviorDetector specifies the type of Misbehavior Detector.
*/
type MisbehaviorDetector uint16

func MisbehaviorDetectorByName(name string) MisbehaviorDetector {
	switch {
	case strings.ToUpper(name) == MBD_DIST_PLAU.String():
		return MBD_DIST_PLAU
	case strings.ToUpper(name) == MBD_SPEE_PLAU.String():
		return MBD_SPEE_PLAU
	case strings.ToUpper(name) == MBD_SPEE_CONS.String():
		return MBD_SPEE_CONS
	case strings.ToUpper(name) == MBD_POS_SPEE_CONS.String():
		return MBD_POS_SPEE_CONS
	case strings.ToUpper(name) == MBD_KALMAN_POS_CONS.String():
		return MBD_KALMAN_POS_CONS
	case strings.ToUpper(name) == MBD_KALMAN_POS_SPEED_CONS_SPEED.String():
		return MBD_KALMAN_POS_SPEED_CONS_SPEED
	case strings.ToUpper(name) == MBD_KALMAN_POS_SPEED_CONS_POS.String():
		return MBD_KALMAN_POS_SPEED_CONS_POS
	case strings.ToUpper(name) == MBD_LOCAL_PERCEPTION_VERIF.String():
		return MBD_LOCAL_PERCEPTION_VERIF
	default:
		return MBD_UNKNOWN
	}

}

func (s MisbehaviorDetector) String() string {
	switch s {
	case MBD_DIST_PLAU:
		return "MBD_DIST_PLAU"
	case MBD_SPEE_PLAU:
		return "MBD_SPEE_PLAU"
	case MBD_SPEE_CONS:
		return "MBD_SPEE_CONS"
	case MBD_POS_SPEE_CONS:
		return "MBD_POS_SPEE_CONS"
	case MBD_KALMAN_POS_CONS:
		return "MBD_KALMAN_POS_CONS"
	case MBD_KALMAN_POS_SPEED_CONS_SPEED:
		return "MBD_KALMAN_POS_SPEED_CONS_SPEED"
	case MBD_KALMAN_POS_SPEED_CONS_POS:
		return "MBD_KALMAN_POS_SPEED_CONS_POS"
	case MBD_LOCAL_PERCEPTION_VERIF:
		return "MBD_LOCAL_PERCEPTION_VERIF"
	default:
		return "UNKNOWN_SOURCE"
	}
}

const (
	MBD_DIST_PLAU MisbehaviorDetector = iota
	MBD_SPEE_PLAU
	MBD_SPEE_CONS
	MBD_POS_SPEE_CONS
	MBD_KALMAN_POS_CONS
	MBD_KALMAN_POS_SPEED_CONS_SPEED
	MBD_KALMAN_POS_SPEED_CONS_POS
	MBD_LOCAL_PERCEPTION_VERIF
	MBD_UNKNOWN
)
