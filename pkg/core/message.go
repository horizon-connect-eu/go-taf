package core

/*
Message represents a wrapper around inbound/outbound messages.
*/
type Message interface {
	Source() string
	Destination() string
	Bytes() []byte
}

type MessageWrapper struct {
	bytes       []byte
	source      string
	destination string
}

func NewMessage(bytes []byte, source string, destination string) Message {
	return &MessageWrapper{
		bytes:       bytes,
		source:      source,
		destination: destination,
	}
}

func (m *MessageWrapper) Source() string {
	return m.source
}

func (m *MessageWrapper) Destination() string {
	return m.destination
}

func (m *MessageWrapper) Bytes() []byte {
	return m.bytes
}
