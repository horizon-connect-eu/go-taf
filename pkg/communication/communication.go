package communication

import (
	"errors"
	"fmt"
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
	return NewWithHandler(tafContext, incomingMessageChannel, outgoingMessageChannel, tafContext.Configuration.CommunicationConfiguration.Handler)
}

func NewWithHandler(tafContext core.RuntimeContext, incomingMessageChannel chan<- Message, outgoingMessageChannel <-chan Message, handlerName string) (CommunicationInterface, error) {

	incomingMessageChannel = make(chan Message, tafContext.Configuration.ChanBufSize)
	outgoingMessageChannel = make(chan Message, tafContext.Configuration.ChanBufSize)

	tafContext.Logger.Warn(fmt.Sprintf("%+v", handlers))

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
