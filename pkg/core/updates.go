package core

/*
The UpdateOp identifies the type of Update Operation than can be applied to an existing TrustModelInstance
*/
type UpdateOp int32

const (
	NO_OP UpdateOp = iota
	UPDATE_TO
	UPDATE_ATO
	ADD_TRUST_OBJECT
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
