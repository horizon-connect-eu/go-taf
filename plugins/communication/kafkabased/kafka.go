package kafkabased

import (
	"context"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/google/uuid"
	logging "github.com/horizon-connect-eu/go-taf/internal/logger"
	"github.com/horizon-connect-eu/go-taf/internal/util"
	"github.com/horizon-connect-eu/go-taf/pkg/communication"
	"github.com/horizon-connect-eu/go-taf/pkg/core"
	"log/slog"
	"os"
	"os/signal"
	"time"
)

func init() {
	communication.RegisterCommunicationHandler("kafka-based", NewKafkaBasedHandler)
}

func NewKafkaBasedHandler(tafContext core.TafContext, inboxChannel chan<- core.Message, outboxChannel <-chan core.Message) {
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

	brokers := []string{tafContext.Configuration.Communication.Kafka.Broker}

	producer, err := sarama.NewAsyncProducer(brokers, config)
	if err != nil {
		logger.Error("Error creating Kafka Producer ", "Details", err)
		os.Exit(-1)
		return
	}
	defer producer.Close()

	//Use randomized ConsumerGroup name to prevent processing of previously missed messages after crash/restart
	consumer, err := sarama.NewConsumerGroup(brokers, tafContext.Identifier+"-"+uuid.New().String(), config)
	if err != nil {
		logger.Error("Error creating Kafka Consumer ", "Details", err)
		os.Exit(-1)
		return

	}
	defer consumer.Close()

	go handleOutgoingMessages(tafContext, logger, producer, outboxChannel)
	go handleIncomingMessages(tafContext, logger, consumer, inboxChannel)

	if err := context.Cause(tafContext.Context); err != nil {
		return
	}
	select {
	case <-tafContext.Context.Done():
		logger.Info("Shutting down Kafka Communication Handler.")
		return
	}
}

func handleOutgoingMessages(tafContext core.TafContext, logger *slog.Logger, producer sarama.AsyncProducer, outboxChannel <-chan core.Message) {
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
				msgAsStr := string(msg.Bytes())
				//logger.Info("Sent message", "Sender", msg.Source(), "Receiving Topic", msg.Destination(), "Message:", msgAsStr, "Offset", success.Offset)
				util.UNUSED(success, msgAsStr)
			case err := <-producer.Errors():
				logger.Error(fmt.Sprintf("Failed to send message: %v", err))
			}
		}
	}
}

func handleIncomingMessages(tafContext core.TafContext, logger *slog.Logger, consumer sarama.ConsumerGroup, inboxChannel chan<- core.Message) {

	ctx, cancel := context.WithCancel(context.Background())
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	for {
		err := consumer.Consume(ctx, []string{tafContext.Configuration.Communication.Kafka.TafTopic}, &consumerHandler{
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
