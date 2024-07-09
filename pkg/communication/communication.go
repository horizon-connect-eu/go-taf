package communication

import (
	"encoding/json"
	"errors"
	"github.com/vs-uulm/go-taf/internal/util"
	"github.com/vs-uulm/go-taf/pkg/command"
	"github.com/vs-uulm/go-taf/pkg/core"
	messages "github.com/vs-uulm/go-taf/pkg/message"
	aivmsg "github.com/vs-uulm/go-taf/pkg/message/aiv"
	mbdmsg "github.com/vs-uulm/go-taf/pkg/message/mbd"
	tasmsg "github.com/vs-uulm/go-taf/pkg/message/tas"
)

var handlers = map[string]CommunicationHandler{}

func RegisterCommunicationHandler(name string, f CommunicationHandler) {
	handlers[name] = f
}

type CommunicationInterface struct {
	tafContext           core.RuntimeContext
	tafChannels          core.TafChannels
	communicationHandler CommunicationHandler
	internalInbox        chan core.Message //message from outside world to CommunicationInterface
	internalOutbox       chan core.Message //message from CommunicationInterface to outside world
}

func NewInterface(tafContext core.RuntimeContext, tafChannels core.TafChannels) (CommunicationInterface, error) {
	return NewInterfaceWithHandler(tafContext, tafChannels, tafContext.Configuration.CommunicationConfiguration.Handler)
}

func NewInterfaceWithHandler(tafContext core.RuntimeContext, tafChannels core.TafChannels, handlerName string) (CommunicationInterface, error) {

	handler, okay := handlers[handlerName]
	if !okay {
		tafContext.Logger.Error("Error creating communication handler '" + handlerName + "'")
		return CommunicationInterface{}, errors.New("Handler " + handlerName + " not found!")
	}

	communicationHandler := CommunicationInterface{
		tafContext:     tafContext,
		tafChannels:    tafChannels,
		internalInbox:  make(chan core.Message, tafContext.Configuration.ChanBufSize),
		internalOutbox: make(chan core.Message, tafContext.Configuration.ChanBufSize),

		communicationHandler: handler,
	}

	return communicationHandler, nil
}

func (ch CommunicationInterface) Run() {

	go ch.communicationHandler(ch.tafContext, ch.internalInbox, ch.internalOutbox)

	go func() {
		for {
			select {
			case msg := <-ch.tafChannels.OutgoingMessageChannel:
				ch.tafContext.Logger.Info("Msg to be sent:", "Msg", string(msg.Bytes()))
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

			msgStr := string(rcvdMsg.Bytes())
			ch.tafContext.Logger.Info("Received message", "Message:", msgStr[0:min(20, len(msgStr)-1)], "Sender", rcvdMsg.Source(), "Topic", rcvdMsg.Destination())

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
				} else {
					cmd := command.CreateTasInitRequest(tasInitReq, rawMsg.Sender, rawMsg.RequestId, rawMsg.ResponseTopic)
					ch.tafChannels.TAMChan <- cmd
				}
			case messages.TAS_TEARDOWN_REQUEST:
				tasTeardownReq, err := tasmsg.UnmarshalTasTeardownRequest(msg)
				if err != nil {
					ch.tafContext.Logger.Error("Error unmarshalling TAS_TEARDOWN_REQUEST: " + err.Error())
				} else {
					cmd := command.CreateTasTeardownRequest(tasTeardownReq, rawMsg.Sender, rawMsg.RequestId, rawMsg.ResponseTopic)
					ch.tafChannels.TAMChan <- cmd
				}
			case messages.TAS_TA_REQUEST:
				tasTaRequest, err := tasmsg.UnmarshalTasTaRequest(msg)
				if err != nil {
					ch.tafContext.Logger.Error("Error unmarshalling TAS_TA_REQUEST: " + err.Error())
				} else {
					util.UNUSED(tasTaRequest) //TODO
				}
			case messages.TAS_SUBSCRIBE_REQUEST:
				tasSubscribeRequest, err := tasmsg.UnmarshalTasSubscribeRequest(msg)
				if err != nil {
					ch.tafContext.Logger.Error("Error unmarshalling TAS_SUBSCRIBE_REQUEST: " + err.Error())
				} else {
					util.UNUSED(tasSubscribeRequest) //TODO
				}
			case messages.TAS_UNSUBSCRIBE_REQUEST:
				tasUnsubscribeRequest, err := tasmsg.UnmarshalTasUnsubscribeRequest(msg)
				if err != nil {
					ch.tafContext.Logger.Error("Error unmarshalling TAS_UNSUBSCRIBE_REQUEST: " + err.Error())
				} else {
					util.UNUSED(tasUnsubscribeRequest) //TODO
				}
			case messages.AIV_RESPONSE:
				aivResponse, err := aivmsg.UnmarshalAivResponse(msg)
				if err != nil {
					ch.tafContext.Logger.Error("Error unmarshalling AIV_RESPONSE: " + err.Error())
				} else {
					util.UNUSED(aivResponse) //TODO
				}
			case messages.AIV_SUBSCRIBE_RESPONSE:
				aivSubscribeResponse, err := aivmsg.UnmarshalAivSubscribeResponse(msg)
				if err != nil {
					ch.tafContext.Logger.Error("Error unmarshalling AIV_SUBSCRIBE_RESPONSE: " + err.Error())
				} else {
					util.UNUSED(aivSubscribeResponse) //TODO
				}
			case messages.AIV_UNSUBSCRIBE_RESPONSE:
				aivUnsubscribeResponse, err := aivmsg.UnmarshalAivUnsubscribeResponse(msg)
				if err != nil {
					ch.tafContext.Logger.Error("Error unmarshalling AIV_UNSUBSCRIBE_RESPONSE: " + err.Error())
				} else {
					util.UNUSED(aivUnsubscribeResponse) //TODO
				}
			case messages.AIV_NOTIFY:
				aivNotify, err := aivmsg.UnmarshalAivNotify(msg)
				if err != nil {
					ch.tafContext.Logger.Error("Error unmarshalling AIV_NOTIFY: " + err.Error())
				} else {
					util.UNUSED(aivNotify) //TODO
				}
			case messages.MBD_SUBSCRIBE_RESPONSE:
				mbdSubscribeResponse, err := mbdmsg.UnmarshalMBDSubscribeResponse(msg)
				if err != nil {
					ch.tafContext.Logger.Error("Error unmarshalling MBD_SUBSCRIBE_RESPONSE: " + err.Error())
				} else {
					util.UNUSED(mbdSubscribeResponse) //TODO
				}
			case messages.MBD_UNSUBSCRIBE_RESPONSE:
				mbdUnsubscribeResponse, err := mbdmsg.UnmarshalMBDUnsubscribeResponse(msg)
				if err != nil {
					ch.tafContext.Logger.Error("Error unmarshalling MBD_UNSUBSCRIBE_RESPONSE: " + err.Error())
				} else {
					util.UNUSED(mbdUnsubscribeResponse) //TODO
				}
			case messages.MBD_NOTIFY:
				mbdNotify, err := mbdmsg.UnmarshalMBDNotify(msg)
				if err != nil {
					ch.tafContext.Logger.Error("Error unmarshalling MBD_NOTIFY: " + err.Error())
				} else {
					util.UNUSED(mbdNotify) //TODO
				}
			case messages.TCH_NOTIFY:
				tchNotify, err := mbdmsg.UnmarshalMBDNotify(msg)
				if err != nil {
					ch.tafContext.Logger.Error("Error unmarshalling TCH_NOTIFY: " + err.Error())
				} else {
					util.UNUSED(tchNotify) //TODO
				}
			default:
				ch.tafContext.Logger.Warn("Received message of type: " + rawMsg.MessageType + ". No processing implemented (yet) for this type of message.")
			}
			//TODO: Add V2X_CAM, V2X_CPM, V2X_NTM

		}
	}
}
