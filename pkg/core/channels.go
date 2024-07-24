package core

type TafChannels struct {
	TAMChannel             chan Command
	OutgoingMessageChannel chan Message
}
