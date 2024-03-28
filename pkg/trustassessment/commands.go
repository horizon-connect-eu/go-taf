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
func CreateInitTMICommand(tmt string, identifier uint64) InitTMICommand {
	return InitTMICommand{
		Identifier:         identifier,
		TrustModelTemplate: tmt,
	}
}

type UpdateTOCommand struct {
	Identifier         uint64
	TrustModelTemplate string
	Trustor            string
	Trustee            string
	TS_ID              string
	Evidence           bool
}

func (receiver UpdateTOCommand) GetType() CommandType {
	return UPDATE_ATO
}
func CreateUpdateTOCommand(identifier uint64, trustor string, trustee string, ts_ID string, evidence bool) UpdateTOCommand {
	return UpdateTOCommand{
		Identifier: identifier,
		Trustor:    trustor,
		Trustee:    trustee,
		TS_ID:      ts_ID,
		Evidence:   evidence,
	}
}
