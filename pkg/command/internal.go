package command

import "github.com/vs-uulm/go-taf/pkg/core"

type HandleTMIUpdate struct {
	commandType core.CommandType
	tmiID       string
	update      core.Update
}

func CreateHandleTMIUpdate(tmiID string, update core.Update) HandleTMIUpdate {
	return HandleTMIUpdate{
		tmiID:       tmiID,
		update:      update,
		commandType: core.HANDLE_TMI_UPDATE,
	}
}

func (r HandleTMIUpdate) Type() core.CommandType {
	return r.commandType
}
