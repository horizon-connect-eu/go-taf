package kafkabased

import (
	logging "github.com/vs-uulm/go-taf/internal/logger"
	"github.com/vs-uulm/go-taf/internal/util"
	"github.com/vs-uulm/go-taf/pkg/communication"
	"github.com/vs-uulm/go-taf/pkg/core"
)

func init() {
	communication.RegisterCommunicationHandler("kafka-based", NewKafkaBasedHandler)
}

func NewKafkaBasedHandler(tafContext core.RuntimeContext, inboxChannel chan<- communication.Message, outboxChannel <-chan communication.Message) {
	logger := logging.CreateChildLogger(tafContext.Logger, "Kafka Communication Handler")
	logger.Info("Starting kafka-based communication handler.")

	//TODO create kafka client, expose producer/consumer to handle functions
}

func handleOutgoingMessages(tafContext core.RuntimeContext, outboxChannel <-chan communication.Message) {
	for {
		select {
		case msg := <-outboxChannel:
			//TODO: send message via Kafka
			util.UNUSED(msg)
		}
	}
}

func handleIncomingMessages(tafContext core.RuntimeContext, inboxChannel chan<- communication.Message) {
	//TODO: read message from Kakfa and put them into channel
}
