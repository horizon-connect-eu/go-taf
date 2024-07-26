package core

/*
The UpdateOp identifies the type of Update Operation than can be applied to an existing TrustModelInstance
*/
type UpdateOp int32

const (
	/*
		Dummy Operation
	*/
	NO_OP UpdateOp = iota
	/*
		Trust Opinion Update
	*/
	UPDATE_TO
	/*
		Atomic Trust Opinion Update
	*/
	UPDATE_ATO
	/*
		Add new Trust Object to Trust Model
	*/
	ADD_TRUST_OBJECT
	/*
		Remove Trust Object from Trust Model
	*/
	REMOVE_TRUST_OBJECT
)

func (u UpdateOp) String() string {
	return [...]string{"NO_OP",
		"UPDATE_TO",
		"UPDATE_ATO",
		"ADD_TRUST_OBJECT",
		"REMOVE_TRUST_OBJECT",
	}[u]
}

/*
An Update operation that must be handled by a Trust Model Instance.
*/
type Update interface {
	Type() UpdateOp
}
