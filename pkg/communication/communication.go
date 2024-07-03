package communication

import (
	"encoding/json"
	"errors"
	"github.com/vs-uulm/go-taf/pkg/command"
	"github.com/vs-uulm/go-taf/pkg/core"
	messages "github.com/vs-uulm/go-taf/pkg/message"
	tasmsg "github.com/vs-uulm/go-taf/pkg/message/tas"
)

var handlers = map[string]CommunicationHandler{}

func RegisterCommunicationHandler(name string, f CommunicationHandler) {
	handlers[name] = f
}

type CommunicationInterface struct {
	tafContext             core.RuntimeContext
	communicationHandler   CommunicationHandler
	outgoingMessageChannel chan Message //message from TAF internals to outer world (proxied by CommunicationInterface)
	internalInbox          chan Message //message from outside world to CommunicationInterface
	internalOutbox         chan Message //message from CommunicationInterface to outside world
}

func New(tafContext core.RuntimeContext, outgoingMessageChannel chan Message) (CommunicationInterface, error) {
	return NewWithHandler(tafContext, outgoingMessageChannel, tafContext.Configuration.CommunicationConfiguration.Handler)
}

func NewWithHandler(tafContext core.RuntimeContext, outgoingMessageChannel chan Message, handlerName string) (CommunicationInterface, error) {

	handler, okay := handlers[handlerName]
	if !okay {
		tafContext.Logger.Error("Error creating communication handler '" + handlerName + "'")
		return CommunicationInterface{}, errors.New("Handler " + handlerName + " not found!")
	}

	communicationHandler := CommunicationInterface{
		tafContext:             tafContext,
		internalInbox:          make(chan Message, tafContext.Configuration.ChanBufSize),
		outgoingMessageChannel: outgoingMessageChannel,
		internalOutbox:         make(chan Message, tafContext.Configuration.ChanBufSize),

		communicationHandler: handler,
	}

	return communicationHandler, nil
}

func (ch CommunicationInterface) Run(tafContext core.RuntimeContext) {
	go ch.communicationHandler(ch.tafContext, ch.internalInbox, ch.internalOutbox)

	go func() {
		for {
			select {
			case msg := <-ch.outgoingMessageChannel:
				ch.tafContext.Logger.Info("Msg rcvd:", "Msg", string(msg.Bytes()))
				ch.internalOutbox <- msg
			}
		}
	}()

	go ch.handleIncomingMessages()

	defer func() {
		//TODO: shutdown
	}()
}

/*
Type with all potential JSON fields of the header structure
*/
type GenericJSONHeaderMessage struct {
	Sender          string
	ServiceType     string
	MessageType     string
	Message         interface{}
	RequestId       string
	ResponseId      string
	ResponseTopic   string
	SubscriberTopic string
}

func (ch CommunicationInterface) handleIncomingMessages() {
	for {
		select {
		case rcvdMsg := <-ch.internalInbox:

			var msg json.RawMessage //Placeholder for the remaining JSON later be unmarshaled using the correct type.
			rawMsg := GenericJSONHeaderMessage{
				Message: &msg,
			}

			//Parse message tpye-agnostically to get type and later unmarshal correct type
			if err := json.Unmarshal(rcvdMsg.Bytes(), &rawMsg); err != nil {
				ch.tafContext.Logger.Error("Error while unmarshalling JSON: " + err.Error())
			}

			schema, exists := messages.SchemaMap[rawMsg.MessageType]
			if !exists {
				ch.tafContext.Logger.Error("Unknown message type: " + rawMsg.MessageType)
			}
			switch schema {
			case messages.TAS_INIT_REQUEST:
				tasInitReq, err := tasmsg.UnmarshalTasInitRequest(msg)
				if err != nil {
					ch.tafContext.Logger.Error("Error unmarshalling TAS_INIT_REQUEST: " + err.Error())
				}
				cmd := command.CreateTasInitRequest(tasInitReq, rawMsg.Sender, rawMsg.RequestId, rawMsg.ResponseTopic)
				ch.tafContext.TAMChan <- cmd
			case messages.TAS_TEARDOWN_REQUEST:
				tasTeardownReq, err := tasmsg.UnmarshalTasTeardownRequest(msg)
				if err != nil {
					ch.tafContext.Logger.Error("Error unmarshalling TAS_TEARDOWN_REQUEST: " + err.Error())
				}
				cmd := command.CreateTasTeardownRequest(tasTeardownReq, rawMsg.Sender, rawMsg.RequestId, rawMsg.ResponseTopic)
				ch.tafContext.TAMChan <- cmd
			}

		}
	}
}
