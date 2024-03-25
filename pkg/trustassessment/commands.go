package trustassessment

type CommandType int64

const (
	UNDEFINED CommandType = iota
	INIT_TMI
	UPDATE_ATO
)

type Command interface {
	GetType() CommandType
}

type InitTMICommand struct {
	Identifier         uint64
	TrustModelTemplate string
}

func (receiver InitTMICommand) GetType() CommandType {
	return INIT_TMI
}
func CreateInitTMICommand(tmt string, identifier uint64) Command {
	return InitTMICommand{
		Identifier:         identifier,
		TrustModelTemplate: tmt,
	}
}

type UpdateATOCommand struct {
	Identifier         uint64
	TrustModelTemplate string
}

func (receiver UpdateATOCommand) GetType() CommandType {
	return UPDATE_ATO
}
func CreateUpdateUpdateATOCommand(tmt string, identifier uint64) Command {
	return UpdateATOCommand{
		Identifier:         identifier,
		TrustModelTemplate: tmt,
	}
}
