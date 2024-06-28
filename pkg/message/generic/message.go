// This file was generated from JSON Schema using quicktype, do not modify it directly.
// To parse and unparse this JSON data, add this code to your project and do:
//
//    genericSubscriptionNotify, err := UnmarshalGenericSubscriptionNotify(bytes)
//    bytes, err = genericSubscriptionNotify.Marshal()
//
//    genericRequest, err := UnmarshalGenericRequest(bytes)
//    bytes, err = genericRequest.Marshal()
//
//    genericResponse, err := UnmarshalGenericResponse(bytes)
//    bytes, err = genericResponse.Marshal()
//
//    genericOneWay, err := UnmarshalGenericOneWay(bytes)
//    bytes, err = genericOneWay.Marshal()
//
//    genericSubscriptionRequest, err := UnmarshalGenericSubscriptionRequest(bytes)
//    bytes, err = genericSubscriptionRequest.Marshal()
//
//    genericSubscriptionResponse, err := UnmarshalGenericSubscriptionResponse(bytes)
//    bytes, err = genericSubscriptionResponse.Marshal()

package genericmsg

import "encoding/json"

func UnmarshalGenericSubscriptionNotify(data []byte) (GenericSubscriptionNotify, error) {
	var r GenericSubscriptionNotify
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *GenericSubscriptionNotify) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalGenericRequest(data []byte) (GenericRequest, error) {
	var r GenericRequest
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *GenericRequest) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalGenericResponse(data []byte) (GenericResponse, error) {
	var r GenericResponse
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *GenericResponse) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalGenericOneWay(data []byte) (GenericOneWay, error) {
	var r GenericOneWay
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *GenericOneWay) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalGenericSubscriptionRequest(data []byte) (GenericSubscriptionRequest, error) {
	var r GenericSubscriptionRequest
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *GenericSubscriptionRequest) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalGenericSubscriptionResponse(data []byte) (GenericSubscriptionResponse, error) {
	var r GenericSubscriptionResponse
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *GenericSubscriptionResponse) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type GenericSubscriptionNotify struct {
	// The actual application message of type messageType.
	Message map[string]interface{} `json:"message"`
	// The message type to be used by the receiver to process this message.
	MessageType string `json:"messageType"`
	// The identifier of the sender of this message.
	Sender string `json:"sender"`
	// The service type to be used by the receiver to process this message.
	ServiceType string `json:"serviceType"`
}

type GenericRequest struct {
	// The actual application message of type messageType.
	Message map[string]interface{} `json:"message"`
	// The message type to be used by the receiver to process this message.
	MessageType string `json:"messageType"`
	// The unique identifier to be repeated in the response for linking request and response.
	RequestID string `json:"requestId"`
	// The Kafka topic to be used to send a response to this request.
	ResponseTopic string `json:"responseTopic"`
	// The identifier of the sender of this message.
	Sender string `json:"sender"`
	// The service type to be used by the receiver to process this message.
	ServiceType string `json:"serviceType"`
}

type GenericResponse struct {
	// The actual application message of type messageType.
	Message map[string]interface{} `json:"message"`
	// The message type to be used by the receiver to process this message.
	MessageType string `json:"messageType"`
	// The unique identifier copied from the request for linking request and response.
	ResponseID string `json:"responseId"`
	// The identifier of the sender of this message.
	Sender string `json:"sender"`
	// The service type to be used by the receiver to process this message.
	ServiceType string `json:"serviceType"`
}

type GenericOneWay struct {
	// The actual application message of type messageType.
	Message map[string]interface{} `json:"message"`
	// The message type to be used by the receiver to process this message.
	MessageType string `json:"messageType"`
	// The identifier of the sender of this message.
	Sender string `json:"sender"`
	// The service type to be used by the receiver to process this message.
	ServiceType string `json:"serviceType"`
}

type GenericSubscriptionRequest struct {
	// The actual application message of type messageType.
	Message map[string]interface{} `json:"message"`
	// The message type to be used by the receiver to process this message.
	MessageType string `json:"messageType"`
	// The unique identifier to be repeated in the response for linking request and response.
	RequestID string `json:"requestId"`
	// The Kafka topic to be used to send a response to this request.
	ResponseTopic string `json:"responseTopic"`
	// The identifier of the sender of this message.
	Sender string `json:"sender"`
	// The service type to be used by the receiver to process this message.
	ServiceType string `json:"serviceType"`
	// The Kafka topic of this sender to which notifications should be published to.
	SubscriberTopic string `json:"subscriberTopic"`
}

type GenericSubscriptionResponse struct {
	// The actual application message of type messageType.
	Message map[string]interface{} `json:"message"`
	// The message type to be used by the receiver to process this message.
	MessageType string `json:"messageType"`
	// The unique identifier to be repeated in the response for linking request and response.
	RequestID string `json:"requestId"`
	// The identifier of the sender of this message.
	Sender string `json:"sender"`
	// The service type to be used by the receiver to process this message.
	ServiceType string `json:"serviceType"`
}
