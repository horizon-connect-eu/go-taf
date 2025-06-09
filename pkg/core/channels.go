package core

/*
TafChannels contains the structs needed by different components to dispatch commands.
*/
type TafChannels struct {
	TAMChannel             chan Command //inbox channel for the TAM
	OutgoingMessageChannel chan Message //inbox channel for outgoing messages, used to dispatch outbound messages
}
