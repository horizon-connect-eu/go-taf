package command

import (
	"github.com/vs-uulm/go-taf/pkg/core"
	tasmsg "github.com/vs-uulm/go-taf/pkg/message/tas"
)

// HandleTasInitRequest  Internal command for handling a received TAS_INIT_REQUEST
type HandleTasInitRequest struct {
	Request       tasmsg.TasInitRequest
	Sender        string
	RequestID     string
	ResponseTopic string
}

func (receiver HandleTasInitRequest) Type() core.CommandType {
	return core.HANDLE_TAS_INIT_REQUEST
}

func CreateTasInitRequest(msg tasmsg.TasInitRequest, sender string, requestID string, responseTopic string) HandleTasInitRequest {
	return HandleTasInitRequest{
		Request:       msg,
		Sender:        sender,
		RequestID:     requestID,
		ResponseTopic: responseTopic,
	}
}

type HandleTasTeardownRequest struct {
	Request       tasmsg.TasTeardownRequest
	Sender        string
	RequestID     string
	ResponseTopic string
}

func (receiver HandleTasTeardownRequest) Type() core.CommandType {
	return core.HANDLE_TAS_TEARDOWN_REQUEST
}

func CreateTasTeardownRequest(msg tasmsg.TasTeardownRequest, sender string, requestID string, responseTopic string) HandleTasTeardownRequest {
	return HandleTasTeardownRequest{
		Request:       msg,
		Sender:        sender,
		RequestID:     requestID,
		ResponseTopic: responseTopic,
	}
}

// HandleTasTaRequest  Internal command for handling a received TAS_TA_REQUEST
type HandleTasTaRequest struct {
	Request       tasmsg.TasTaRequest
	Sender        string
	RequestID     string
	ResponseTopic string
}

func (receiver HandleTasTaRequest) Type() core.CommandType {
	return core.HANDLE_TAS_TA_REQUEST
}

func CreateHandleTasTaRequest(msg tasmsg.TasTaRequest, sender string, requestID string, responseTopic string) HandleTasTaRequest {
	return HandleTasTaRequest{
		Request:       msg,
		Sender:        sender,
		RequestID:     requestID,
		ResponseTopic: responseTopic,
	}
}
