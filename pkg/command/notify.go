package command

import (
	"github.com/vs-uulm/go-taf/pkg/core"
	aivmsg "github.com/vs-uulm/go-taf/pkg/message/aiv"
	mbdmsg "github.com/vs-uulm/go-taf/pkg/message/mbd"
	tchmsg "github.com/vs-uulm/go-taf/pkg/message/tch"
)

type NotifyMessage interface {
	aivmsg.AivNotify | mbdmsg.MBDNotify | tchmsg.TchNotify
}

type HandleNotify[R NotifyMessage] struct {
	Notify      R
	Sender      string
	commandType core.CommandType
}

func (r HandleNotify[notify]) Type() core.CommandType {
	return r.commandType
}

func CreateAivNotify(msg aivmsg.AivNotify, sender string) HandleNotify[aivmsg.AivNotify] {
	return HandleNotify[aivmsg.AivNotify]{
		Notify:      msg,
		Sender:      sender,
		commandType: core.HANDLE_AIV_NOTIFY,
	}
}

func CreateMbdNotify(msg mbdmsg.MBDNotify, sender string) HandleNotify[mbdmsg.MBDNotify] {
	return HandleNotify[mbdmsg.MBDNotify]{
		Notify:      msg,
		Sender:      sender,
		commandType: core.HANDLE_MBD_NOTIFY,
	}
}

func CreateTchNotify(msg tchmsg.TchNotify, sender string) HandleNotify[tchmsg.TchNotify] {
	return HandleNotify[tchmsg.TchNotify]{
		Notify:      msg,
		Sender:      sender,
		commandType: core.HANDLE_TCH_NOTIFY,
	}
}
