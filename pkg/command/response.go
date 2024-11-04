package command

import (
	"github.com/vs-uulm/go-taf/pkg/core"
	aivmsg "github.com/vs-uulm/go-taf/pkg/message/aiv"
	mbdmsg "github.com/vs-uulm/go-taf/pkg/message/mbd"
	taqimsg "github.com/vs-uulm/go-taf/pkg/message/taqi"
)

type response interface {
	aivmsg.AivResponse | taqimsg.TaqiResult | subscriptionResponse
}

type subscriptionResponse interface {
	aivmsg.AivSubscribeResponse | aivmsg.AivUnsubscribeResponse | mbdmsg.MBDSubscribeResponse | mbdmsg.MBDUnsubscribeResponse
}

type HandleResponse[R response] struct {
	Response    R
	Sender      string
	ResponseID  string
	commandType core.CommandType
}

func (r HandleResponse[response]) Type() core.CommandType {
	return r.commandType
}

func CreateAivResponse(msg aivmsg.AivResponse, sender string, responseID string) HandleResponse[aivmsg.AivResponse] {
	return HandleResponse[aivmsg.AivResponse]{
		Response:    msg,
		Sender:      sender,
		ResponseID:  responseID,
		commandType: core.HANDLE_AIV_RESPONSE,
	}
}

func CreateAivSubscriptionResponse(msg aivmsg.AivSubscribeResponse, sender string, responseID string) HandleResponse[aivmsg.AivSubscribeResponse] {
	return HandleResponse[aivmsg.AivSubscribeResponse]{
		Response:    msg,
		Sender:      sender,
		ResponseID:  responseID,
		commandType: core.HANDLE_AIV_SUBSCRIBE_RESPONSE,
	}
}

func CreateAivUnsubscriptionResponse(msg aivmsg.AivUnsubscribeResponse, sender string, responseID string) HandleResponse[aivmsg.AivUnsubscribeResponse] {
	return HandleResponse[aivmsg.AivUnsubscribeResponse]{
		Response:    msg,
		Sender:      sender,
		ResponseID:  responseID,
		commandType: core.HANDLE_AIV_UNSUBSCRIBE_RESPONSE,
	}
}

func CreateMbdSubscriptionResponse(msg mbdmsg.MBDSubscribeResponse, sender string, responseID string) HandleResponse[mbdmsg.MBDSubscribeResponse] {
	return HandleResponse[mbdmsg.MBDSubscribeResponse]{
		Response:    msg,
		Sender:      sender,
		ResponseID:  responseID,
		commandType: core.HANDLE_MBD_SUBSCRIBE_RESPONSE,
	}
}

func CreateMbdUnsubscriptionResponse(msg mbdmsg.MBDUnsubscribeResponse, sender string, responseID string) HandleResponse[mbdmsg.MBDUnsubscribeResponse] {
	return HandleResponse[mbdmsg.MBDUnsubscribeResponse]{
		Response:    msg,
		Sender:      sender,
		ResponseID:  responseID,
		commandType: core.HANDLE_MBD_UNSUBSCRIBE_RESPONSE,
	}
}

func CreateTaqiResult(msg taqimsg.TaqiResult, sender string, responseID string) HandleResponse[taqimsg.TaqiResult] {
	return HandleResponse[taqimsg.TaqiResult]{
		Response:    msg,
		Sender:      sender,
		ResponseID:  responseID,
		commandType: core.HANDLE_TAQI_RESULT,
	}
}
