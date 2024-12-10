package core

import "bytes"

/*
The UpdateOp identifies the type of Update Operation than can be applied to an existing TrustModelInstance.
*/
type UpdateOp uint16

const (
	/*
		Dummy Operation
	*/
	NO_OP UpdateOp = iota
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
	/*
		Update a Trust Model based upon a CPM information
	*/
	REFRESH_CPM
)

func (u UpdateOp) String() string {
	return [...]string{
		"NO_OP",
		"UPDATE_ATO",
		"ADD_TRUST_OBJECT",
		"REMOVE_TRUST_OBJECT",
		"REFRESH_CPM",
	}[u]
}

/*
An Update operation that must be handled by a Trust Model Instance.
*/
type Update interface {
	Type() UpdateOp
}

func (u *UpdateOp) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(u.String())
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}
