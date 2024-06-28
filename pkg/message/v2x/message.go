// This file was generated from JSON Schema using quicktype, do not modify it directly.
// To parse and unparse this JSON data, add this code to your project and do:
//
//    v2XCpm, err := UnmarshalV2XCpm(bytes)
//    bytes, err = v2XCpm.Marshal()
//
//    v2XNtm, err := UnmarshalV2XNtm(bytes)
//    bytes, err = v2XNtm.Marshal()

package v2xmsg

import "encoding/json"

func UnmarshalV2XCpm(data []byte) (V2XCpm, error) {
	var r V2XCpm
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *V2XCpm) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalV2XNtm(data []byte) (V2XNtm, error) {
	var r V2XNtm
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *V2XNtm) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type V2XCpm struct {
	Latitude                   float64                  `json:"latitude"`
	Longitude                  float64                  `json:"longitude"`
	OrientationAngle           float64                  `json:"orientationAngle"`
	OrientationAngleConfidence float64                  `json:"orientationAngleConfidence"`
	PerceivedObjectContainer   PerceivedObjectContainer `json:"PerceivedObjectContainer"`
	ReferenceTime              float64                  `json:"referenceTime"`
	SemiMajorConfidence        float64                  `json:"semiMajorConfidence"`
	SemiMajorOrientation       float64                  `json:"semiMajorOrientation"`
	SemiMinorConfidence        float64                  `json:"semiMinorConfidence"`
	SourceID                   float64                  `json:"sourceId"`
}

type PerceivedObjectContainer struct {
	NumberOfPerceivedObjects float64  `json:"numberOfPerceivedObjects"`
	Objects                  []Object `json:"objects"`
}

type Object struct {
	AccelerationDirection                float64 `json:"accelerationDirection"`
	AccelerationDirectionConfidence      float64 `json:"accelerationDirectionConfidence"`
	AccelerationMagnitudeValue           float64 `json:"accelerationMagnitudeValue"`
	AccelerationMagnitudeValueConfidence float64 `json:"accelerationMagnitudeValueConfidence"`
	MeasurementDeltaTime                 float64 `json:"measurementDeltaTime"`
	ObjectID                             float64 `json:"objectId"`
	PositionXCoordinate                  float64 `json:"position_xCoordinate"`
	PositionXCoordinateConfidence        float64 `json:"position_xCoordinateConfidence"`
	PositionYCoordinate                  float64 `json:"position_yCoordinate"`
	PositionYCoordinateConfidence        float64 `json:"position_yCoordinateConfidence"`
	VelocityDirection                    float64 `json:"velocityDirection"`
	VelocityDirectionConfidence          float64 `json:"velocityDirectionConfidence"`
	VelocityMagnitudeValue               float64 `json:"velocityMagnitudeValue"`
	VelocityMagnitudeValueConfidence     float64 `json:"velocityMagnitudeValueConfidence"`
	ZAngle                               float64 `json:"zAngle"`
	ZAngleConfidence                     float64 `json:"zAngleConfidence"`
	ZAngularVelocity                     float64 `json:"zAngularVelocity"`
	ZAngularVelocityConfidence           float64 `json:"zAngularVelocityConfidence"`
}

type V2XNtm struct {
	V2XSourceSet []V2XSourceSet `json:"v2xSourceSet"`
}

type V2XSourceSet struct {
	Opinion     Opinion `json:"opinion"`
	V2XSourceID int64   `json:"v2xSourceId"`
}

type Opinion struct {
	BaseRate    float64 `json:"baseRate"`
	Belief      float64 `json:"belief"`
	Disbelief   float64 `json:"disbelief"`
	Uncertainty float64 `json:"uncertainty"`
}
