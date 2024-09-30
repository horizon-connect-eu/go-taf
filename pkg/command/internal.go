package command

import (
	"github.com/vs-uulm/go-taf/pkg/core"
)

type HandleTMIUpdate struct {
	commandType core.CommandType
	FullTmiID   string
	Updates     []core.Update
}

func CreateHandleTMIUpdate(FullTmiID string, updates ...core.Update) HandleTMIUpdate {
	return HandleTMIUpdate{
		FullTmiID:   FullTmiID,
		Updates:     updates,
		commandType: core.HANDLE_TMI_UPDATE,
	}
}

func (r HandleTMIUpdate) Type() core.CommandType {
	return r.commandType
}

type HandleTMIInit struct {
	commandType core.CommandType
	FullTMI     string
	TMI         core.TrustModelInstance
}

func CreateHandleTMIInit(fullTMIid string, TMI core.TrustModelInstance) HandleTMIInit {
	return HandleTMIInit{

		TMI:         TMI,
		FullTMI:     fullTMIid,
		commandType: core.HANDLE_TMI_INIT,
	}
}

func (r HandleTMIInit) Type() core.CommandType {
	return r.commandType
}

type HandleTMIDestroy struct {
	commandType core.CommandType
	FullTMI     string
}

func CreateHandleTMIDestroy(fullTMIid string) HandleTMIDestroy {
	return HandleTMIDestroy{
		FullTMI:     fullTMIid,
		commandType: core.HANDLE_TMI_DESTROY,
	}
}

func (r HandleTMIDestroy) Type() core.CommandType {
	return r.commandType
}

type HandleATLUpdate struct {
	commandType core.CommandType
	FullTMI     string
	ResultSet   core.AtlResultSet
}

func CreateHandleATLUpdate(atl core.AtlResultSet, fullTMI string) HandleATLUpdate {
	return HandleATLUpdate{
		ResultSet:   atl,
		FullTMI:     fullTMI,
		commandType: core.HANDLE_ATL_UPDATE,
	}
}

func (r HandleATLUpdate) Type() core.CommandType {
	return r.commandType
}
