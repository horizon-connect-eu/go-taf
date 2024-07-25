package communication

import (
	"encoding/json"
)

type GenericRequestWrapper struct {
	Sender        string      `json:"sender"`
	ServiceType   string      `json:"serviceType"`
	MessageType   string      `json:"messageType"`
	ResponseTopic string      `json:"responseTopic"`
	RequestId     string      `json:"requestId"`
	Message       interface{} `json:"message"`
}

type GenericResponseWrapper struct {
	Sender      string      `json:"sender"`
	ServiceType string      `json:"serviceType"`
	MessageType string      `json:"messageType"`
	ResponseId  string      `json:"responseId"`
	Message     interface{} `json:"message"`
}

type GenericSubscriptionRequestWrapper struct {
	Sender          string      `json:"sender"`
	ServiceType     string      `json:"serviceType"`
	MessageType     string      `json:"messageType"`
	ResponseTopic   string      `json:"responseTopic"`
	SubscriberTopic string      `json:"subscriberTopic"`
	RequestId       string      `json:"requestId"`
	Message         interface{} `json:"message"`
}

/*
Function builds a byte representation of a JSON response by filling the header fields and
*/
func BuildRequest(sender string, serviceType string, messageType string, responseTopic string, message interface{}) ([]byte, error) {
	responseWrapper := GenericRequestWrapper{
		Sender:        sender,
		ServiceType:   serviceType,
		MessageType:   messageType,
		RequestId:     generateRequestId(),
		ResponseTopic: responseTopic,
		Message:       message,
	}
	return json.Marshal(responseWrapper)
}

func generateRequestId() string {
	return "123" //TODO
}

/*
Function builds a byte representation of a JSON response by filling the header fields and
*/
func BuildSubscriptionRequest(sender string, serviceType string, messageType string, responseTopic string, subscriberTopic string, message interface{}) ([]byte, error) {
	subReqWrapper := GenericSubscriptionRequestWrapper{
		Sender:          sender,
		ServiceType:     serviceType,
		MessageType:     messageType,
		RequestId:       generateRequestId(),
		SubscriberTopic: subscriberTopic,
		ResponseTopic:   responseTopic,
		Message:         message,
	}
	return json.Marshal(subReqWrapper)
}

/*
Function builds a byte representation of a JSON response by filling the header fields and
*/
func BuildResponse(sender string, serviceType string, messageType string, responseId string, message interface{}) ([]byte, error) {
	responseWrapper := GenericResponseWrapper{
		Sender:      sender,
		ServiceType: serviceType,
		MessageType: messageType,
		ResponseId:  responseId,
		Message:     message,
	}
	return json.Marshal(responseWrapper)
}
