package communication

import (
	"errors"
	"github.com/vs-uulm/go-taf/pkg/core"
)

var handlers = map[string]CommunicationHandler{}

func RegisterCommunicationHandler(name string, f CommunicationHandler) {
	handlers[name] = f
}

type CommunicationInterface struct {
	tafContext             core.RuntimeContext
	communicationHandler   CommunicationHandler
	incomingMessageChannel chan<- Message
	outgoingMessageChannel <-chan Message
}

func New(tafContext core.RuntimeContext, incomingMessageChannel chan<- Message, outgoingMessageChannel <-chan Message) (CommunicationInterface, error) {

	handlerName := tafContext.Configuration.CommunicationConfiguration.Handler
	handler, okay := handlers[handlerName]
	if !okay {
		tafContext.Logger.Error("Error creating communication handler '" + handlerName + "'")
		return CommunicationInterface{}, errors.New("Handler " + handlerName + " not found!")
	}

	communicationHandler := CommunicationInterface{
		tafContext:             tafContext,
		incomingMessageChannel: incomingMessageChannel,
		outgoingMessageChannel: outgoingMessageChannel,
		communicationHandler:   handler,
	}

	return communicationHandler, nil
}

func (ch CommunicationInterface) Run(tafContext core.RuntimeContext) {
	go ch.communicationHandler(ch.tafContext, ch.incomingMessageChannel, ch.outgoingMessageChannel)

	defer func() {
		//TODO: shutdown
	}()
}
