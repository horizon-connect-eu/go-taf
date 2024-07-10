package command

import (
	"github.com/vs-uulm/go-taf/pkg/core"
	tasmsg "github.com/vs-uulm/go-taf/pkg/message/tas"
)

type request interface {
	tasmsg.TasInitRequest | tasmsg.TasTeardownRequest | tasmsg.TasTaRequest
}

type subscriptionRequest interface {
	tasmsg.TasSubscribeRequest | tasmsg.TasUnsubscribeRequest
}

type HandleRequest[R request] struct {
	Request       R
	Sender        string
	RequestID     string
	ResponseTopic string
	commandType   core.CommandType
}

type HandleSubscriptionRequest[R subscriptionRequest] struct {
	Request         R
	Sender          string
	RequestID       string
	ResponseTopic   string
	SubscriberTopic string
	commandType     core.CommandType
}

func (r HandleRequest[tasRequest]) Type() core.CommandType {
	return r.commandType
}
func (r HandleSubscriptionRequest[tasRequest]) Type() core.CommandType {
	return r.commandType
}

func CreateTasInitRequest(msg tasmsg.TasInitRequest, sender string, requestID string, responseTopic string) HandleRequest[tasmsg.TasInitRequest] {
	return HandleRequest[tasmsg.TasInitRequest]{
		Request:       msg,
		Sender:        sender,
		RequestID:     requestID,
		ResponseTopic: responseTopic,
		commandType:   core.HANDLE_TAS_INIT_REQUEST,
	}
}

func CreateTasTaRequest(msg tasmsg.TasTaRequest, sender string, requestID string, responseTopic string) HandleRequest[tasmsg.TasTaRequest] {
	return HandleRequest[tasmsg.TasTaRequest]{
		Request:       msg,
		Sender:        sender,
		RequestID:     requestID,
		ResponseTopic: responseTopic,
		commandType:   core.HANDLE_TAS_TA_REQUEST,
	}
}

func CreateTasTeardownRequest(msg tasmsg.TasTeardownRequest, sender string, requestID string, responseTopic string) HandleRequest[tasmsg.TasTeardownRequest] {
	return HandleRequest[tasmsg.TasTeardownRequest]{
		Request:       msg,
		Sender:        sender,
		RequestID:     requestID,
		ResponseTopic: responseTopic,
		commandType:   core.HANDLE_TAS_TEARDOWN_REQUEST,
	}
}

func CreateTasSubscribeRequest(msg tasmsg.TasSubscribeRequest, sender string, requestID string, responseTopic string, subscriberTopic string) HandleSubscriptionRequest[tasmsg.TasSubscribeRequest] {
	return HandleSubscriptionRequest[tasmsg.TasSubscribeRequest]{
		Request:         msg,
		Sender:          sender,
		RequestID:       requestID,
		ResponseTopic:   responseTopic,
		SubscriberTopic: subscriberTopic,
		commandType:     core.HANDLE_TAS_SUBSCRIBE_REQUEST,
	}
}

func CreateTasUnsubscribeRequest(msg tasmsg.TasUnsubscribeRequest, sender string, requestID string, responseTopic string, subscriberTopic string) HandleSubscriptionRequest[tasmsg.TasUnsubscribeRequest] {
	return HandleSubscriptionRequest[tasmsg.TasUnsubscribeRequest]{
		Request:         msg,
		Sender:          sender,
		RequestID:       requestID,
		ResponseTopic:   responseTopic,
		SubscriberTopic: subscriberTopic,
		commandType:     core.HANDLE_TAS_UNSUBSCRIBE_REQUEST,
	}
}

/*

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


*/
