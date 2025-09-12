package communication

import (
	"encoding/json"
	messages "github.com/horizon-connect-eu/go-taf/pkg/message"
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

type GenericSubscriptionResponseWrapper struct {
	Sender      string      `json:"sender"`
	ServiceType string      `json:"serviceType"`
	MessageType string      `json:"messageType"`
	ResponseId  string      `json:"responseId"`
	Message     interface{} `json:"message"`
}

type GenericOneWayMessageWrapper struct {
	Sender      string      `json:"sender"`
	ServiceType string      `json:"serviceType"`
	MessageType string      `json:"messageType"`
	Message     interface{} `json:"message"`
}

/*
The BuildRequest function builds a byte representation of a JSON request by filling the header fields and return a byte representation of the message and the request ID used.
*/
func BuildRequest(sender string, messageType messages.MessageSchema, responseTopic string, requestId string, message interface{}) ([]byte, error) {
	responseWrapper := GenericRequestWrapper{
		Sender:        sender,
		ServiceType:   messages.ServiceMap[messageType],
		MessageType:   string(messageType),
		RequestId:     requestId,
		ResponseTopic: responseTopic,
		Message:       message,
	}
	bytes, err := json.Marshal(responseWrapper)
	if err != nil {
		return nil, err
	} else {
		return bytes, nil
	}
}

/*
The BuildSubscriptionRequest function builds a byte representation of a JSON subscription request by filling the header fields and returns a byte representation of the message and the request ID used.
*/
func BuildSubscriptionRequest(sender string, messageType messages.MessageSchema, responseTopic string, subscriberTopic string, requestId string, message interface{}) ([]byte, error) {
	subReqWrapper := GenericSubscriptionRequestWrapper{
		Sender:          sender,
		ServiceType:     messages.ServiceMap[messageType],
		MessageType:     string(messageType),
		RequestId:       requestId,
		SubscriberTopic: subscriberTopic,
		ResponseTopic:   responseTopic,
		Message:         message,
	}
	bytes, err := json.Marshal(subReqWrapper)
	if err != nil {
		return nil, err
	} else {
		return bytes, nil
	}

}

/*
The BuildResponse function builds a byte representation of a JSON response by filling the header fields and returns a byte representation of the message.
*/
func BuildResponse(sender string, messageType messages.MessageSchema, responseId string, message interface{}) ([]byte, error) {
	responseWrapper := GenericResponseWrapper{
		Sender:      sender,
		ServiceType: messages.ServiceMap[messageType],
		MessageType: string(messageType),
		ResponseId:  responseId,
		Message:     message,
	}
	return json.Marshal(responseWrapper)
}

/*
The BuildSubscriptionResponse function builds a byte representation of a JSON subscription response by filling the header fields and returns a byte representation of the message.
*/
func BuildSubscriptionResponse(sender string, messageType messages.MessageSchema, responseId string, message interface{}) ([]byte, error) {
	subResWrapper := GenericSubscriptionResponseWrapper{
		Sender:      sender,
		ServiceType: messages.ServiceMap[messageType],
		MessageType: string(messageType),
		ResponseId:  responseId,
		Message:     message,
	}
	return json.Marshal(subResWrapper)
}

/*
The BuildOneWayMessage function builds a byte representation of a JSON subscription response by filling the header fields and returns a byte representation of the message.
*/
func BuildOneWayMessage(sender string, messageType messages.MessageSchema, message interface{}) ([]byte, error) {
	msgWrapper := GenericOneWayMessageWrapper{
		Sender:      sender,
		ServiceType: messages.ServiceMap[messageType],
		MessageType: string(messageType),
		Message:     message,
	}
	return json.Marshal(msgWrapper)
}
