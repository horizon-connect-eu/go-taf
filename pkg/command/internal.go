package command

import (
	"github.com/vs-uulm/go-taf/pkg/core"
)

type HandleTMIUpdate struct {
	commandType core.CommandType
	TmiID       string
	Update      []core.Update
}

func CreateHandleTMIUpdate(tmiID string, update ...core.Update) HandleTMIUpdate {
	return HandleTMIUpdate{
		TmiID:       tmiID,
		Update:      update,
		commandType: core.HANDLE_TMI_UPDATE,
	}
}

func (r HandleTMIUpdate) Type() core.CommandType {
	return r.commandType
}

type HandleTMIInit struct {
	commandType core.CommandType
	TmiID       string
	SessionID   string
	TMI         core.TrustModelInstance
}

func CreateHandleTMIInit(tmiID string, TMI core.TrustModelInstance, SessionID string) HandleTMIInit {
	return HandleTMIInit{

		TmiID:       tmiID,
		TMI:         TMI,
		SessionID:   SessionID,
		commandType: core.HANDLE_TMI_INIT,
	}
}

func (r HandleTMIInit) Type() core.CommandType {
	return r.commandType
}

type HandleTMIDestroy struct {
	commandType core.CommandType
	TmiID       string
}

func CreateHandleTMIDestroy(tmiID string) HandleTMIDestroy {
	return HandleTMIDestroy{
		TmiID:       tmiID,
		commandType: core.HANDLE_TMI_DESTROY,
	}
}

func (r HandleTMIDestroy) Type() core.CommandType {
	return r.commandType
}

type HandleATLUpdate struct {
	commandType core.CommandType
	Session     string
	ResultSet   core.AtlResultSet
}

func CreateHandleATLUpdate(atl core.AtlResultSet, session string) HandleATLUpdate {
	return HandleATLUpdate{
		ResultSet:   atl,
		Session:     session,
		commandType: core.HANDLE_ATL_UPDATE,
	}
}

func (r HandleATLUpdate) Type() core.CommandType {
	return r.commandType
}
