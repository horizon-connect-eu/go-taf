package kafkabased

import (
	"context"
	"fmt"
	"github.com/IBM/sarama"
	logging "github.com/vs-uulm/go-taf/internal/logger"
	"github.com/vs-uulm/go-taf/pkg/communication"
	"github.com/vs-uulm/go-taf/pkg/core"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"time"
)

func init() {
	communication.RegisterCommunicationHandler("kafka-based", NewKafkaBasedHandler)
}

func NewKafkaBasedHandler(tafContext core.RuntimeContext, inboxChannel chan<- core.Message, outboxChannel <-chan core.Message) {
	logger := logging.CreateChildLogger(tafContext.Logger, "Kafka Communication Handler")
	logger.Info("Starting kafka-based communication handler.")

	config := sarama.NewConfig()
	config.Version = sarama.V2_1_0_0
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	config.Consumer.Offsets.AutoCommit.Enable = true
	config.Consumer.Offsets.AutoCommit.Interval = 1 * time.Second
	config.Producer.RequiredAcks = sarama.WaitForLocal
	config.Producer.Return.Errors = true
	config.Producer.Return.Successes = true

	brokers := []string{tafContext.Configuration.CommunicationConfiguration.Kafka.Broker}

	producer, err := sarama.NewAsyncProducer(brokers, config)
	if err != nil {
		logger.Error("Error creating Kafka Producer ", "Details", err)
		return
	}
	defer producer.Close()

	consumer, err := sarama.NewConsumerGroup(brokers, tafContext.Identifier, config)
	if err != nil {
		logger.Error("Error creating Kafka Consumer ", "Details", err)
		return

	}
	defer consumer.Close()

	var wg sync.WaitGroup
	wg.Add(1)

	go handleOutgoingMessages(tafContext, logger, producer, outboxChannel)
	go handleIncomingMessages(tafContext, logger, consumer, inboxChannel)

	wg.Wait() //TODO: fix for orderly shutdown
}

func handleOutgoingMessages(tafContext core.RuntimeContext, logger *slog.Logger, producer sarama.AsyncProducer, outboxChannel <-chan core.Message) {
	for {
		select {
		case msg := <-outboxChannel:

			kafkaMsg := &sarama.ProducerMessage{
				Topic: msg.Destination(),
				Value: sarama.ByteEncoder(msg.Bytes()),
			}

			producer.Input() <- kafkaMsg

			select {
			case success := <-producer.Successes():
				//				logger.Info("Message sent", "Message:", string(msg.Bytes()), "Offset", success.Offset)
				msgAsStr := string(msg.Bytes())
				logger.Info("Message sent", "Sender", msg.Source(), "Receiving Topic", msg.Destination(), "Message Excerpt:", msgAsStr[0:min(20, len(msgAsStr)-1)], "Offset", success.Offset)
			case err := <-producer.Errors():
				logger.Error(fmt.Sprintf("Failed to send message: %v", err))
			}
		}
	}
}

func handleIncomingMessages(tafContext core.RuntimeContext, logger *slog.Logger, consumer sarama.ConsumerGroup, inboxChannel chan<- core.Message) {

	//TODO: fix context usage
	ctx, cancel := context.WithCancel(context.Background())
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	for {
		err := consumer.Consume(ctx, tafContext.Configuration.CommunicationConfiguration.Kafka.Topics, &consumerHandler{
			inboxChannel: inboxChannel,
			logger:       logger,
		})
		if err != nil {
			logger.Error(fmt.Sprintf("consume error: %v", err))
		}

		select {
		case <-signals:
			cancel()
			return
		default:
		}
	}

}

type consumerHandler struct {
	inboxChannel chan<- core.Message
	logger       *slog.Logger
}

func (h *consumerHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		//convert Kafka message to internally wrapped message
		internalMsg := core.NewMessage(msg.Value, "", msg.Topic)
		h.inboxChannel <- internalMsg
		sess.MarkMessage(msg, "")
	}
	return nil
}
