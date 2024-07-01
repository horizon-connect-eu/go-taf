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
	incomingMessageChannel chan<- Message //message from outside world to TAF internals (proxied by CommunicationInterface)
	outgoingMessageChannel <-chan Message //message from TAF internals to outer world (proxied by CommunicationInterface)
	internalInbox          <-chan Message //message from outside world to CommunicationInterface
	internalOutbox         chan<- Message //message from CommunicationInterface to outside world
}

func New(tafContext core.RuntimeContext, incomingMessageChannel chan<- Message, outgoingMessageChannel <-chan Message) (CommunicationInterface, error) {
	return NewWithHandler(tafContext, incomingMessageChannel, outgoingMessageChannel, tafContext.Configuration.CommunicationConfiguration.Handler)
}

func NewWithHandler(tafContext core.RuntimeContext, incomingMessageChannel chan<- Message, outgoingMessageChannel <-chan Message, handlerName string) (CommunicationInterface, error) {

	tafContext.Logger.Warn(fmt.Sprintf("%+v", handlers))

	handler, okay := handlers[handlerName]
	if !okay {
		tafContext.Logger.Error("Error creating communication handler '" + handlerName + "'")
		return CommunicationInterface{}, errors.New("Handler " + handlerName + " not found!")
	}

	communicationHandler := CommunicationInterface{
		tafContext:             tafContext,
		incomingMessageChannel: incomingMessageChannel,
		internalInbox:          make(chan Message, tafContext.Configuration.ChanBufSize),
		outgoingMessageChannel: outgoingMessageChannel,
		internalOutbox:         make(chan Message, tafContext.Configuration.ChanBufSize),

		communicationHandler: handler,
	}

	return communicationHandler, nil
}

func (ch CommunicationInterface) Run(tafContext core.RuntimeContext) {
	go ch.communicationHandler(ch.tafContext, ch.incomingMessageChannel, ch.outgoingMessageChannel)

	go func() {
		for {
			select {
			case msg := <-ch.outgoingMessageChannel:
				ch.internalOutbox <- msg
			}
		}
	}()

	go func() {
		for {
			select {
			case msg := <-ch.internalInbox:
				ch.incomingMessageChannel <- msg
			}
		}

	}()

	defer func() {
		//TODO: shutdown
	}()
}
