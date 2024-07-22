package command

import (
	"github.com/vs-uulm/go-taf/pkg/core"
	v2xmsg "github.com/vs-uulm/go-taf/pkg/message/v2x"
)

type oneway interface {
	v2xmsg.V2XCpm | v2xmsg.V2XNtm
}

type HandleOneWay[R oneway] struct {
	OneWay      R
	Sender      string
	commandType core.CommandType
}

func (r HandleOneWay[oneway]) Type() core.CommandType {
	return r.commandType
}

func CreateV2xCpm(msg v2xmsg.V2XCpm, sender string) HandleOneWay[v2xmsg.V2XCpm] {
	return HandleOneWay[v2xmsg.V2XCpm]{
		OneWay:      msg,
		Sender:      sender,
		commandType: core.HANDLE_V2X_CPM,
	}
}

func CreateV2xNtm(msg v2xmsg.V2XNtm, sender string) HandleOneWay[v2xmsg.V2XNtm] {
	return HandleOneWay[v2xmsg.V2XNtm]{
		OneWay:      msg,
		Sender:      sender,
		commandType: core.HANDLE_V2X_NTM,
	}
}
