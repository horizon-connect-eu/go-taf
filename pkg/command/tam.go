package command

import tasmsg "github.com/vs-uulm/go-taf/pkg/message/tas"

type HandleTasInitRequest struct {
	msg           tasmsg.TasInitRequest
	sender        string
	requestID     string
	responseTopic string
}

func (receiver HandleTasInitRequest) Type() CommandType {
	return HANDLE_TAS_INIT_REQUEST
}

func CreateTasInitRequest(msg tasmsg.TasInitRequest, sender string, requestID string, responseTopic string) HandleTasInitRequest {
	return HandleTasInitRequest{
		msg:           msg,
		sender:        sender,
		requestID:     requestID,
		responseTopic: responseTopic,
	}
}
func (receiver HandleTasInitRequest) Request() tasmsg.TasInitRequest {
	return receiver.msg
}

func (receiver HandleTasInitRequest) Sender() string {
	return receiver.sender
}

func (receiver HandleTasInitRequest) RequestID() string {
	return receiver.requestID
}

func (receiver HandleTasInitRequest) ResponseTopic() string {
	return receiver.responseTopic
}
