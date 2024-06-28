package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	logging "github.com/vs-uulm/go-taf/internal/logger"
	"github.com/vs-uulm/go-taf/internal/validator"
	"github.com/vs-uulm/go-taf/pkg/config"
	message2 "github.com/vs-uulm/go-taf/pkg/message"
	aivmsg "github.com/vs-uulm/go-taf/pkg/message/aiv"
	mbdmsg "github.com/vs-uulm/go-taf/pkg/message/mbd"
	tasmsg "github.com/vs-uulm/go-taf/pkg/message/tas"
	v2xmsg "github.com/vs-uulm/go-taf/pkg/message/v2x"
	"log"
	"log/slog"
	"math/rand/v2"
	"os"
	"os/signal"
	"sync"
	"time"
)

var WATCH_TOPICS = []string{"taf", "tch", "aiv", "mbd", "application.ccam"}
var logger *slog.Logger

/*
A helper command to watch and check Kafka topics
*/
func main() {
	tafConfig := config.DefaultConfig
	// First, see whether a config file path has been specified
	if filepath, ok := os.LookupEnv("TAF_CONFIG"); ok {
		var err error
		tafConfig, err = config.LoadJSON(filepath)
		if err != nil {
			log.Fatalf("main: error reading config file %s: %s\n", filepath, err.Error())
		}
	}

	logger = logging.CreateMainLogger(tafConfig.Logging)
	logger.Info("Configuration loaded")
	logger.Debug("Running with following configuration",
		slog.String("CONFIG", fmt.Sprintf("%+v", tafConfig)))

	saramaConsume()

	/*

	 */

}

func saramaConsume() {
	config := sarama.NewConfig()
	config.Version = sarama.V2_1_0_0
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	config.Consumer.Offsets.AutoCommit.Enable = true
	config.Consumer.Offsets.AutoCommit.Interval = 1 * time.Second

	brokers := []string{"localhost:9092"}

	client, err := sarama.NewConsumerGroup(brokers, "cg"+string(rand.IntN(1000000)), config)
	if err != nil {
		logger.Warn(fmt.Sprintf("unable to create kafka consumer group: %v", err))
	}
	defer client.Close()

	ctx, cancel := context.WithCancel(context.Background())
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	var wg sync.WaitGroup
	wg.Add(1)

	handler := &consumerHandler{}

	go func() {
		defer wg.Done()

		for {
			err := client.Consume(ctx, WATCH_TOPICS, handler)
			if err != nil {
				logger.Warn(fmt.Sprintf("Error from consumer: %v", err))
			}

			select {
			case <-signals:
				cancel()
				return
			default:
			}
		}
	}()

	wg.Wait()
}

type consumerHandler struct {
}

func (h *consumerHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		logger.Info("Received message:",
			slog.String("Topic", string(msg.Topic)),
			slog.String("Offset", string(msg.Offset)),
			slog.String("Partition", string(msg.Partition)),
			slog.String("Key", string(msg.Key)),
			slog.String("Value", string(msg.Value)),
		)
		sess.MarkMessage(msg, "")
		checkMessage(string(msg.Value))
	}
	return nil
}

type GenericMessage struct {
	ServiceType string
	MessageType string
	Message     interface{}
}

func checkMessage(message string) {
	var msg json.RawMessage //Placeholder for the remaining JSON later be unmarshaled using the correct type.
	rawMsg := GenericMessage{
		Message: &msg,
	}

	//Parse message tpye-agnostically to get type and later unmarshal correct type
	if err := json.Unmarshal([]byte(message), &rawMsg); err != nil {
		logger.Error(err.Error())
	}

	schema, exists := message2.SchemaMap[rawMsg.MessageType]
	if !exists {
		logger.Error("Unknown message type: " + rawMsg.MessageType)
	} else {
		valid, w, err := validator.Validate(schema, string(msg))
		if err != nil {
			logger.Error(err.Error())
		} else if !valid {
			logger.Error("Error validating document", "Errors", w)
		} else {
			logger.Info("Successfully validated JSON with schema " + string(schema) + ".")

			var err error
			switch schema {
			case message2.AIV_NOTIFY:
				_, err = aivmsg.UnmarshalAivNotify(msg)
			case message2.AIV_REQUEST:
				_, err = aivmsg.UnmarshalAivRequest(msg)
			case message2.AIV_RESPONSE:
				_, err = aivmsg.UnmarshalAivResponse(msg)
			case message2.AIV_SUBSCRIBE_REQUEST:
				_, err = aivmsg.UnmarshalAivSubscribeRequest(msg)
			case message2.AIV_SUBSCRIBE_RESPONSE:
				_, err = aivmsg.UnmarshalAivSubscribeResponse(msg)
			case message2.AIV_UNSUBSCRIBE_REQUEST:
				_, err = aivmsg.UnmarshalAivUnsubscribeRequest(msg)
			case message2.AIV_UNSUBSCRIBE_RESPONSE:
				_, err = aivmsg.UnmarshalAivUnsubscribeResponse(msg)
			case message2.MBD_NOTIFY:
				_, err = mbdmsg.UnmarshalMBDNotify(msg)
			case message2.MBD_SUBSCRIBE_REQUEST:
				_, err = mbdmsg.UnmarshalMBDSubscribeRequest(msg)
			case message2.MBD_SUBSCRIBE_RESPONSE:
				_, err = mbdmsg.UnmarshalMBDSubscribeResponse(msg)
			case message2.MBD_UNSUBSCRIBE_REQUEST:
				_, err = mbdmsg.UnmarshalMBDUnsubscribeRequest(msg)
			case message2.MBD_UNSUBSCRIBE_RESPONSE:
				_, err = mbdmsg.UnmarshalMBDUnsubscribeResponse(msg)
			case message2.TAS_INIT_REQUEST:
				_, err = tasmsg.UnmarshalTasInitRequest(msg)
			case message2.TAS_INIT_RESPONSE:
				_, err = tasmsg.UnmarshalTasInitResponse(msg)
			case message2.TAS_NOTIFY:
				_, err = tasmsg.UnmarshalTasNotify(msg)
			case message2.TAS_SUBSCRIBE_REQUEST:
				_, err = tasmsg.UnmarshalTasSubscribeRequest(msg)
			case message2.TAS_SUBSCRIBE_RESPONSE:
				_, err = tasmsg.UnmarshalTasSubscribeResponse(msg)
			case message2.TAS_TA_REQUEST:
				_, err = tasmsg.UnmarshalTasTaRequest(msg)
			case message2.TAS_TA_RESPONSE:
				_, err = tasmsg.UnmarshalTasTaResponse(msg)
			case message2.TAS_TEARDOWN_REQUEST:
				_, err = tasmsg.UnmarshalTasTeardownRequest(msg)
			case message2.TAS_TEARDOWN_RESPONSE:
				_, err = tasmsg.UnmarshalTasTeardownResponse(msg)
			case message2.TAS_UNSUBSCRIBE_REQUEST:
				_, err = tasmsg.UnmarshalTasUnsubscribeRequest(msg)
			case message2.TAS_UNSUBSCRIBE_RESPONSE:
				_, err = tasmsg.UnmarshalTasUnsubscribeResponse(msg)
			case message2.V2X_CPM:
				_, err = v2xmsg.UnmarshalV2XCpm(msg)
			case message2.V2X_NTM:
				_, err = v2xmsg.UnmarshalV2XNtm(msg)
			}
			if err != nil {
				logger.Error(err.Error())
			} else {
				logger.Info("Successfully unmarshalled struct of type " + string(schema) + ".")
			}
		}
	}

	//{  "sender": "a77b29bac8f1-taf",  "serviceType": "TAS",  "messageType": "TAS_INIT_REQUEST",  "responseId": "4c54a50f8e43",  "message" : {  "trustModelTemplate":"TRUSTMODEL@0.0.1"}}

}
