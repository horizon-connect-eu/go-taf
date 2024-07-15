package communication

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/vs-uulm/go-taf/internal/util"
	"github.com/vs-uulm/go-taf/pkg/command"
	"github.com/vs-uulm/go-taf/pkg/core"
	messages "github.com/vs-uulm/go-taf/pkg/message"
	aivmsg "github.com/vs-uulm/go-taf/pkg/message/aiv"
	mbdmsg "github.com/vs-uulm/go-taf/pkg/message/mbd"
	tasmsg "github.com/vs-uulm/go-taf/pkg/message/tas"
	tchmsg "github.com/vs-uulm/go-taf/pkg/message/tch"
	v2xmsg "github.com/vs-uulm/go-taf/pkg/message/v2x"
	"strings"
)

var handlers = map[string]CommunicationHandler{}

func RegisterCommunicationHandler(name string, f CommunicationHandler) {
	handlers[name] = f
}

type CommunicationInterface struct {
	tafContext           core.TafContext
	channels             core.TafChannels
	communicationHandler CommunicationHandler
	internalInbox        chan core.Message //message from outside world to CommunicationInterface
	internalOutbox       chan core.Message //message from CommunicationInterface to outside world
}

func NewInterface(tafContext core.TafContext, tafChannels core.TafChannels) (CommunicationInterface, error) {
	return NewInterfaceWithHandler(tafContext, tafChannels, tafContext.Configuration.CommunicationConfiguration.Handler)
}

func NewInterfaceWithHandler(tafContext core.TafContext, tafChannels core.TafChannels, handlerName string) (CommunicationInterface, error) {

	handler, okay := handlers[handlerName]
	if !okay {
		tafContext.Logger.Error("Error creating communication handler '" + handlerName + "'")
		return CommunicationInterface{}, errors.New("Handler " + handlerName + " not found!")
	}

	communicationHandler := CommunicationInterface{
		tafContext:     tafContext,
		channels:       tafChannels,
		internalInbox:  make(chan core.Message, tafContext.Configuration.ChanBufSize),
		internalOutbox: make(chan core.Message, tafContext.Configuration.ChanBufSize),

		communicationHandler: handler,
	}

	return communicationHandler, nil
}

func (ch CommunicationInterface) Run() {
	defer func() {
		ch.tafContext.Logger.Info("Shutting down Communication Interface.")
	}()

	go ch.communicationHandler(ch.tafContext, ch.internalInbox, ch.internalOutbox)

	go func() {
		for {
			if err := context.Cause(ch.tafContext.Context); err != nil {
				return
			}
			select {
			case <-ch.tafContext.Context.Done():
				return
			case msg := <-ch.channels.OutgoingMessageChannel:
				ch.tafContext.Logger.Info("Msg to be sent:", "Msg", string(msg.Bytes()))
				ch.internalOutbox <- msg
			}
		}
	}()

	go ch.handleIncomingMessages()

	for {
		// Each iteration, check whether we've been cancelled.
		if err := context.Cause(ch.tafContext.Context); err != nil {
			return
		}
		select {
		case <-ch.tafContext.Context.Done():
			return

		}
	}

}

// GenericJSONHeaderMessage Type with all potential JSON fields of the header structure
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
				} else if ok, errs := checkRequestFields(rawMsg); !ok {
					ch.tafContext.Logger.Error("Incomplete message header for TAS_INIT_REQUEST message: " + errs.Error())
				} else {
					cmd := command.CreateTasInitRequest(tasInitReq, rawMsg.Sender, rawMsg.RequestId, rawMsg.ResponseTopic)
					ch.channels.TAMChan <- cmd
				}
			case messages.TAS_TEARDOWN_REQUEST:
				tasTeardownReq, err := tasmsg.UnmarshalTasTeardownRequest(msg)
				if err != nil {
					ch.tafContext.Logger.Error("Error unmarshalling TAS_TEARDOWN_REQUEST: " + err.Error())
				} else if ok, errs := checkRequestFields(rawMsg); !ok {
					ch.tafContext.Logger.Error("Incomplete message header for TAS_TEARDOWN_REQUEST message: " + errs.Error())
				} else {
					cmd := command.CreateTasTeardownRequest(tasTeardownReq, rawMsg.Sender, rawMsg.RequestId, rawMsg.ResponseTopic)
					ch.channels.TAMChan <- cmd
				}
			case messages.TAS_TA_REQUEST:
				tasTaRequest, err := tasmsg.UnmarshalTasTaRequest(msg)
				if err != nil {
					ch.tafContext.Logger.Error("Error unmarshalling TAS_TA_REQUEST: " + err.Error())
				} else if ok, errs := checkRequestFields(rawMsg); !ok {
					ch.tafContext.Logger.Error("Incomplete message header for TAS_TA_REQUEST message: " + errs.Error())
				} else {
					cmd := command.CreateTasTaRequest(tasTaRequest, rawMsg.Sender, rawMsg.RequestId, rawMsg.ResponseTopic)
					ch.channels.TAMChan <- cmd
				}
			case messages.TAS_SUBSCRIBE_REQUEST:
				tasSubscribeRequest, err := tasmsg.UnmarshalTasSubscribeRequest(msg)
				if err != nil {
					ch.tafContext.Logger.Error("Error unmarshalling TAS_SUBSCRIBE_REQUEST: " + err.Error())
				} else if ok, errs := checkSubscriptionRequestFields(rawMsg); !ok {
					ch.tafContext.Logger.Error("Incomplete message header for TAS_SUBSCRIBE_REQUEST message: " + errs.Error())
				} else {
					cmd := command.CreateTasSubscribeRequest(tasSubscribeRequest, rawMsg.Sender, rawMsg.RequestId, rawMsg.ResponseTopic, rawMsg.SubscriberTopic)
					ch.channels.TAMChan <- cmd
				}
			case messages.TAS_UNSUBSCRIBE_REQUEST:
				tasUnsubscribeRequest, err := tasmsg.UnmarshalTasUnsubscribeRequest(msg)
				if err != nil {
					ch.tafContext.Logger.Error("Error unmarshalling TAS_UNSUBSCRIBE_REQUEST: " + err.Error())
				} else if ok, errs := checkSubscriptionRequestFields(rawMsg); !ok {
					ch.tafContext.Logger.Error("Incomplete message header for TAS_UNSUBSCRIBE_REQUEST message: " + errs.Error())
				} else {
					cmd := command.CreateTasUnsubscribeRequest(tasUnsubscribeRequest, rawMsg.Sender, rawMsg.RequestId, rawMsg.ResponseTopic, rawMsg.SubscriberTopic)
					ch.channels.TAMChan <- cmd
				}
			case messages.AIV_RESPONSE:
				aivResponse, err := aivmsg.UnmarshalAivResponse(msg)
				if err != nil {
					ch.tafContext.Logger.Error("Error unmarshalling AIV_RESPONSE: " + err.Error())
				} else if ok, errs := checkResponseFields(rawMsg); !ok {
					ch.tafContext.Logger.Error("Incomplete message header for AIV_RESPONSE message: " + errs.Error())
				} else {
					cmd := command.CreateAivResponse(aivResponse, rawMsg.Sender, rawMsg.ResponseId)
					ch.channels.TSMChan <- cmd
				}
			case messages.AIV_SUBSCRIBE_RESPONSE:
				aivSubscribeResponse, err := aivmsg.UnmarshalAivSubscribeResponse(msg)
				if err != nil {
					ch.tafContext.Logger.Error("Error unmarshalling AIV_SUBSCRIBE_RESPONSE: " + err.Error())
				} else if ok, errs := checkSubscriptionResponseFields(rawMsg); !ok {
					ch.tafContext.Logger.Error("Incomplete message header for AIV_SUBSCRIBE_RESPONSE message: " + errs.Error())
				} else {
					cmd := command.CreateAivSubscriptionResponse(aivSubscribeResponse, rawMsg.Sender, rawMsg.ResponseId)
					util.UNUSED(cmd) //TODO
				}
			case messages.AIV_UNSUBSCRIBE_RESPONSE:
				aivUnsubscribeResponse, err := aivmsg.UnmarshalAivUnsubscribeResponse(msg)
				if err != nil {
					ch.tafContext.Logger.Error("Error unmarshalling AIV_UNSUBSCRIBE_RESPONSE: " + err.Error())
				} else if ok, errs := checkSubscriptionResponseFields(rawMsg); !ok {
					ch.tafContext.Logger.Error("Incomplete message header for AIV_UNSUBSCRIBE_RESPONSE message: " + errs.Error())
				} else {
					cmd := command.CreateAivUnsubscriptionResponse(aivUnsubscribeResponse, rawMsg.Sender, rawMsg.ResponseId)
					util.UNUSED(cmd) //TODO
				}
			case messages.AIV_NOTIFY:
				aivNotify, err := aivmsg.UnmarshalAivNotify(msg)
				if err != nil {
					ch.tafContext.Logger.Error("Error unmarshalling AIV_NOTIFY: " + err.Error())
				} else if ok, errs := checkNotifyFields(rawMsg); !ok {
					ch.tafContext.Logger.Error("Incomplete message header for AIV_NOTIFY message: " + errs.Error())
				} else {
					cmd := command.CreateAivNotify(aivNotify, rawMsg.Sender)
					ch.channels.TSMChan <- cmd
				}
			case messages.MBD_SUBSCRIBE_RESPONSE:
				mbdSubscribeResponse, err := mbdmsg.UnmarshalMBDSubscribeResponse(msg)
				if err != nil {
					ch.tafContext.Logger.Error("Error unmarshalling MBD_SUBSCRIBE_RESPONSE: " + err.Error())
				} else if ok, errs := checkSubscriptionResponseFields(rawMsg); !ok {
					ch.tafContext.Logger.Error("Incomplete message header for MBD_SUBSCRIBE_RESPONSE message: " + errs.Error())
				} else {
					cmd := command.CreateMbdSubscriptionResponse(mbdSubscribeResponse, rawMsg.Sender, rawMsg.ResponseId)
					util.UNUSED(cmd) //TODO
				}
			case messages.MBD_UNSUBSCRIBE_RESPONSE:
				mbdUnsubscribeResponse, err := mbdmsg.UnmarshalMBDUnsubscribeResponse(msg)
				if err != nil {
					ch.tafContext.Logger.Error("Error unmarshalling MBD_UNSUBSCRIBE_RESPONSE: " + err.Error())
				} else if ok, errs := checkSubscriptionResponseFields(rawMsg); !ok {
					ch.tafContext.Logger.Error("Incomplete message header for MBD_UNSUBSCRIBE_RESPONSE message: " + errs.Error())
				} else {
					cmd := command.CreateMbdUnsubscriptionResponse(mbdUnsubscribeResponse, rawMsg.Sender, rawMsg.ResponseId)
					util.UNUSED(cmd) //TODO
				}
			case messages.MBD_NOTIFY:
				mbdNotify, err := mbdmsg.UnmarshalMBDNotify(msg)
				if err != nil {
					ch.tafContext.Logger.Error("Error unmarshalling MBD_NOTIFY: " + err.Error())
				} else if ok, errs := checkNotifyFields(rawMsg); !ok {
					ch.tafContext.Logger.Error("Incomplete message header for MBD_NOTIFY message: " + errs.Error())
				} else {
					cmd := command.CreateMbdNotify(mbdNotify, rawMsg.Sender)
					util.UNUSED(cmd) //TODO
				}
			case messages.TCH_NOTIFY:
				tchNotify, err := tchmsg.UnmarshalMessage(msg)
				if err != nil {
					ch.tafContext.Logger.Error("Error unmarshalling TCH_NOTIFY: " + err.Error())
				} else if ok, errs := checkNotifyFields(rawMsg); !ok {
					ch.tafContext.Logger.Error("Incomplete message header for TCH_NOTIFY message: " + errs.Error())
				} else {
					cmd := command.CreateTchNotify(tchNotify, rawMsg.Sender)
					util.UNUSED(cmd) //TODO
				}
			case messages.V2X_NTM:
				v2xNtm, err := v2xmsg.UnmarshalV2XNtm(msg)
				if err != nil {
					ch.tafContext.Logger.Error("Error unmarshalling V2X_NTM: " + err.Error())
				} else {
					util.UNUSED(v2xNtm) //TODO
				}
			case messages.V2X_CPM:
				v2xCpm, err := v2xmsg.UnmarshalV2XCpm(msg)
				if err != nil {
					ch.tafContext.Logger.Error("Error unmarshalling V2X_CPM: " + err.Error())
				} else {
					util.UNUSED(v2xCpm) //TODO
				}
			default:
				ch.tafContext.Logger.Warn("Received message of type: " + rawMsg.MessageType + ". No processing implemented (yet) for this type of message.")
			}
		}
	}
}

// Takes a raw message and checks whether required fields are set for GENERIC_REQUEST messages.
func checkRequestFields(msg GenericJSONHeaderMessage) (bool, error) {
	errs := make([]string, 0, 3)
	if len(msg.Sender) == 0 {
		errs = append(errs, "Sender field is empty.")
	}
	if len(msg.RequestId) == 0 {
		errs = append(errs, "Request ID is missing.")
	}
	if len(msg.ResponseTopic) == 0 {
		errs = append(errs, "Response Topic is missing.")
	}
	if len(errs) > 0 {
		return false, errors.New(strings.Join(errs, "\n"))
	} else {
		return true, nil
	}
}

// Takes a raw message and checks whether required fields are set for GENERIC_RESPONSE messages.
func checkResponseFields(msg GenericJSONHeaderMessage) (bool, error) {
	errs := make([]string, 0, 2)
	if len(msg.Sender) == 0 {
		errs = append(errs, "Sender field is empty.")
	}
	if len(msg.ResponseId) == 0 {
		errs = append(errs, "Response ID is missing.")
	}
	if len(errs) > 0 {
		return false, errors.New(strings.Join(errs, "\n"))
	} else {
		return true, nil
	}
}

// Takes a raw message and checks whether required fields are set for GENERIC_SUBSCRIPTION_REQUEST messages.
func checkSubscriptionRequestFields(msg GenericJSONHeaderMessage) (bool, error) {
	errs := make([]string, 0, 4)
	if len(msg.Sender) == 0 {
		errs = append(errs, "Sender field is empty.")
	}
	if len(msg.RequestId) == 0 {
		errs = append(errs, "Request ID is missing.")
	}
	if len(msg.ResponseTopic) == 0 {
		errs = append(errs, "Response Topic is missing.")
	}
	if len(msg.SubscriberTopic) == 0 {
		errs = append(errs, "Subscription Topic is missing.")
	}
	if len(errs) > 0 {
		return false, errors.New(strings.Join(errs, "\n"))
	} else {
		return true, nil
	}
}

// Takes a raw message and checks whether required fields are set for GENERIC_SUBSCRIPTION_RESPONSE messages.
func checkSubscriptionResponseFields(msg GenericJSONHeaderMessage) (bool, error) {
	errs := make([]string, 0, 2)
	if len(msg.Sender) == 0 {
		errs = append(errs, "Sender field is empty.")
	}
	if len(msg.ResponseId) == 0 {
		errs = append(errs, "Response ID is missing.")
	}
	if len(errs) > 0 {
		return false, errors.New(strings.Join(errs, "\n"))
	} else {
		return true, nil
	}
}

// Takes a raw message and checks whether required fields are set for GENERIC_REQUEST messages.
func checkNotifyFields(msg GenericJSONHeaderMessage) (bool, error) {
	errs := make([]string, 0, 1)
	if len(msg.Sender) == 0 {
		errs = append(errs, "Sender field is empty.")
	}
	if len(errs) > 0 {
		return false, errors.New(strings.Join(errs, "\n"))
	} else {
		return true, nil
	}
}
