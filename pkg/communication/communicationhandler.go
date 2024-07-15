package communication

import "github.com/vs-uulm/go-taf/pkg/core"

/*
The function to be implemented by communication handlers.
The outgoingMessageChannel needs to be filled by the Communcation Handler once external messages have arrived at the TAF.
In turn, the incomingMessageChannel queues messages from the TAF that need to be sent to external components.
*/
type CommunicationHandler func(tafContext core.TafContext, incomingMessageChannel chan<- core.Message, outgoingMessageChannel <-chan core.Message)
