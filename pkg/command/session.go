package command

import (
	"github.com/vs-uulm/go-taf/pkg/core"
	tasmsg "github.com/vs-uulm/go-taf/pkg/message/tas"
)

type HandleTasInitRequest struct {
	msg           tasmsg.TasInitRequest
	sender        string
	requestID     string
	responseTopic string
}

func (receiver HandleTasInitRequest) Type() core.CommandType {
	return core.HANDLE_TAS_INIT_REQUEST
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

type HandleTasTeardownRequest struct {
	msg           tasmsg.TasTeardownRequest
	sender        string
	requestID     string
	responseTopic string
}

func (receiver HandleTasTeardownRequest) Type() core.CommandType {
	return core.HANDLE_TAS_TEARDOWN_REQUEST
}

func CreateTasTeardownRequest(msg tasmsg.TasTeardownRequest, sender string, requestID string, responseTopic string) HandleTasTeardownRequest {
	return HandleTasTeardownRequest{
		msg:           msg,
		sender:        sender,
		requestID:     requestID,
		responseTopic: responseTopic,
	}
}
func (receiver HandleTasTeardownRequest) Request() tasmsg.TasTeardownRequest {
	return receiver.msg
}

func (receiver HandleTasTeardownRequest) Sender() string {
	return receiver.sender
}

func (receiver HandleTasTeardownRequest) RequestID() string {
	return receiver.requestID
}

func (receiver HandleTasTeardownRequest) ResponseTopic() string {
	return receiver.responseTopic
}
