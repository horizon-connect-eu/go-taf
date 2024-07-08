package core

type TafChannels struct {
	TAMChan                chan Command
	TSMChan                chan Command
	TMMChan                chan Command
	OutgoingMessageChannel chan Message
}
