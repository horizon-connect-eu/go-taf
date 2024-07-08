package communication

import "encoding/json"

type genericRequestWrapper struct {
	sender        string      `json:"sender"`
	serviceType   string      `json:"serviceType"`
	messageType   string      `json:"messageType"`
	responseTopic string      `json:"responseTopic"`
	requestId     string      `json:"requestId"`
	message       interface{} `json:"message"`
}

type genericResponseWrapper struct {
	sender      string      `json:"sender"`
	serviceType string      `json:"serviceType"`
	messageType string      `json:"messageType"`
	responseId  string      `json:"responseId"`
	message     interface{} `json:"message"`
}

/*
Function builds a byte representation of a JSON response by filling the header fields and
*/
func BuildRequest(sender string, serviceType string, messageType string, responseTopic string, message interface{}) ([]byte, error) {
	responseWrapper := genericRequestWrapper{
		sender:        sender,
		serviceType:   serviceType,
		messageType:   messageType,
		requestId:     generateRequestId(),
		responseTopic: responseTopic,
		message:       message,
	}
	return json.Marshal(responseWrapper)
}

func generateRequestId() string {
	return "123" //TODO
}

/*
Function builds a byte representation of a JSON response by filling the header fields and
*/
func BuildResponse(sender string, serviceType string, messageType string, responseId string, message interface{}) ([]byte, error) {
	responseWrapper := genericResponseWrapper{
		sender:      sender,
		serviceType: serviceType,
		messageType: messageType,
		responseId:  responseId,
		message:     message,
	}
	return json.Marshal(responseWrapper)
}
