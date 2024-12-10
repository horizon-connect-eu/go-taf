package listener

import "bytes"

type EventType uint8

const (
	ATL_UPDATED EventType = iota
	ATL_REMOVED
	SESSION_CREATED
	SESSION_TORNDOWN
	TRUST_MODEL_INSTANCE_SPAWNED
	TRUST_MODEL_INSTANCE_UPDATED
	TRUST_MODEL_INSTANCE_DELETED
)

func (e EventType) String() string {
	return [...]string{"ATL_UPDATED", "ATL_REMOVED", "SESSION_CREATED", "SESSION_TORNDOWN", "TRUST_MODEL_INSTANCE_SPAWNED", "TRUST_MODEL_INSTANCE_UPDATED", "TRUST_MODEL_INSTANCE_DELETED"}[e]
}

func (e EventType) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(e.String())
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

type ListenerEvent interface {
	Event() EventType
}
