package filebased

import (
	"fmt"
	logging "github.com/vs-uulm/go-taf/internal/logger"
	"github.com/vs-uulm/go-taf/pkg/communication"
	"github.com/vs-uulm/go-taf/pkg/core"
)

func init() {
	communication.RegisterCommunicationHandler("file-based", NewFileBasedHandler)
}

func NewFileBasedHandler(tafContext core.RuntimeContext, inboxChannel chan<- communication.Message, outboxChannel <-chan communication.Message) {
	logger := logging.CreateChildLogger(tafContext.Logger, "File Communication Handler")
	logger.Info("Starting file-based communication handler.")

	go handleOutgoingMessages(tafContext, outboxChannel)
	go handleIncomingMessages(tafContext, inboxChannel)
}

/*
Print message content to console.
*/
func handleOutgoingMessages(tafContext core.RuntimeContext, outboxChannel <-chan communication.Message) {
	for {
		select {
		case msg := <-outboxChannel:
			fmt.Printf("Outgoing message from %s to %s:", msg.Source(), msg.Destination())
			fmt.Println(string(msg.Bytes()))
		}
	}
}

func handleIncomingMessages(tafContext core.RuntimeContext, inboxChannel chan<- communication.Message) {
	//TODO: read local messages from local file and send them into inbox
}
