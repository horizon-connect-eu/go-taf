package core

type CommandType int64

const (
	UNDEFINED CommandType = iota
	HANDLE_TAS_INIT_REQUEST
	HANDLE_TAS_TEARDOWN_REQUEST
)

type Command interface {
	Type() CommandType
}
