package command

import (
	"github.com/vs-uulm/go-taf/pkg/core"
)

/*
HandleTMIUpdate contains 1 or more update operations to be applied on a specified Trust Model Instance.
In case of more update operations, a worker should apply micro-batching and apply all updates before proceeding (e.g., calling the TLEE).
*/
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

/*
HandleTMIInit is a command that initiates the existence of a Trust Model Instance for a TAM worker.
*/
type HandleTMIInit struct {
	commandType core.CommandType
	FullTmiID   string
	TMI         core.TrustModelInstance
}

func CreateHandleTMIInit(fullTMIid string, TMI core.TrustModelInstance) HandleTMIInit {
	return HandleTMIInit{

		TMI:         TMI,
		FullTmiID:   fullTMIid,
		commandType: core.HANDLE_TMI_INIT,
	}
}

func (r HandleTMIInit) Type() core.CommandType {
	return r.commandType
}

/*
HandleTMIInit is a command that signals a TAM worker to destroy a Trust Model Instance from its shard.
*/
type HandleTMIDestroy struct {
	commandType core.CommandType
	FullTmiID   string
}

func CreateHandleTMIDestroy(fullTMIid string) HandleTMIDestroy {
	return HandleTMIDestroy{
		FullTmiID:   fullTMIid,
		commandType: core.HANDLE_TMI_DESTROY,
	}
}

func (r HandleTMIDestroy) Type() core.CommandType {
	return r.commandType
}

/*
HandleATLUpdate is a command sent from a TAM worker to the TAM that contains new ATL results to be cached by the TAM.
*/

type HandleATLUpdate struct {
	commandType core.CommandType
	FullTmiID   string
	ResultSet   core.AtlResultSet
}

func CreateHandleATLUpdate(atl core.AtlResultSet, fullTMI string) HandleATLUpdate {
	return HandleATLUpdate{
		ResultSet:   atl,
		FullTmiID:   fullTMI,
		commandType: core.HANDLE_ATL_UPDATE,
	}
}

func (r HandleATLUpdate) Type() core.CommandType {
	return r.commandType
}
