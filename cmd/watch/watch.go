package main

/*
Example messages to input into the CLI KAFAK producer:

{  "sender": "a77b29bac8f1-taf",  "serviceType": "TAS",  "messageType": "TAS_INIT_REQUEST",  "responseId": "4c54a50f8e43",  "message" : {  "trustModelTemplate":"TRUSTMODEL@0.0.1"}}
{  "sender": "a77b29bac8f1-taf",  "serviceType": "TAS",  "messageType": "TAS_INIT_REQUEST",  "responseId": "4c54a50f8e43",  "message" : {  "trustModelTemplateee":"TRUSTMODEL@0.0.1"}}

{  "sender": "a77b29bac8f1-aiv",  "serviceType": "ECI",  "messageType": "AIV_RESPONSE",  "responseId": "4c54a50f8e42",  "message" : {"trusteeReports": [{"trusteeID": "Zonal Controller 1","attestationReport": [{"claim": "secure-boot-integrity","timestamp": "2024-05-16T15:30:45Z","appraisal": 1},{"claim": "runtime-integrity","timestamp": "2024-05-16T15:35:22Z","appraisal": 0}]}],"aivEvidence": {    "timestamp": "2024-05-16T15:30:45Z",    "nonce": "d78080092edf3633e6933f67ddfe6744",    "signatureAlgorithmType": "ECDSA-SHA256",    "signature":"30440220655e8f8b6f96a6c3a21257aab77c1e5c13ae8acf94dabc6b6e13416d2ff3477a022033d5d7dab3f516cd7367e8d637ab4b956aaa080f3c236b78edbd7f5c2ca1a86767",    "keyRef":"ecdsa_public_key_71"  }}}

*/
import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	logging "github.com/vs-uulm/go-taf/internal/logger"
	"github.com/vs-uulm/go-taf/internal/validator"
	"github.com/vs-uulm/go-taf/pkg/config"
	messages "github.com/vs-uulm/go-taf/pkg/message"
	aivmsg "github.com/vs-uulm/go-taf/pkg/message/aiv"
	mbdmsg "github.com/vs-uulm/go-taf/pkg/message/mbd"
	tasmsg "github.com/vs-uulm/go-taf/pkg/message/tas"
	v2xmsg "github.com/vs-uulm/go-taf/pkg/message/v2x"
	"log"
	"log/slog"
	"math/rand/v2"
	"os"
	"os/signal"
	"reflect"
	"sync"
	"syscall"
	"time"
)

var WATCH_TOPICS = []string{"taf", "tch", "aiv", "mbd", "application.ccam"}
var logger *slog.Logger

// Blocks until the process receives SIGTERM (or equivalent).
func WaitForCtrlC() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}

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

	go saramaConsume(tafConfig.CommunicationConfiguration.Kafka)

	WaitForCtrlC()

}

/*
 * The functions registers for the WATCH_TOPICS at the Kafka broker and checks every message it consumes.
 */
func saramaConsume(kafkaConfig config.KafkaConfig) {
	config := sarama.NewConfig()
	config.Version = sarama.V2_1_0_0
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	config.Consumer.Offsets.AutoCommit.Enable = true
	config.Consumer.Offsets.AutoCommit.Interval = 1 * time.Second

	brokers := []string{kafkaConfig.Broker}

	client, err := sarama.NewConsumerGroup(brokers, "cg"+fmt.Sprint(rand.IntN(1000000)), config)
	if err != nil {
		logger.Error(fmt.Sprintf("unable to create kafka consumer group: %v", err))
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
				logger.Error(fmt.Sprintf("Error from consumer: %v", err))
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
			slog.Int64("Offset", msg.Offset),
			slog.Int("Partition", int(msg.Partition)),
			slog.String("Key", string(msg.Key)),
			slog.String("Value", string(msg.Value)),
		)
		checkMessage(string(msg.Value))
		sess.MarkMessage(msg, "")
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
		logger.Error("Error while unmarshalling JSON: " + err.Error())
	}

	schema, exists := messages.SchemaMap[rawMsg.MessageType]
	if !exists {
		logger.Error("Unknown message type: " + rawMsg.MessageType)
	} else {
		valid, w, err := validator.Validate(schema, string(msg))
		if err != nil {
			logger.Error("Error while trying to validate: " + err.Error())
		} else if !valid {
			logger.Error("Error validating document", "Errors", w)
		} else {
			logger.Info("Successfully validated JSON with schema " + string(schema) + ".")

			var err error
			var extractedStruct interface{}

			switch schema {
			case messages.AIV_NOTIFY:
				extractedStruct, err = aivmsg.UnmarshalAivNotify(msg)
			case messages.AIV_REQUEST:
				extractedStruct, err = aivmsg.UnmarshalAivRequest(msg)
			case messages.AIV_RESPONSE:
				extractedStruct, err = aivmsg.UnmarshalAivResponse(msg)
			case messages.AIV_SUBSCRIBE_REQUEST:
				extractedStruct, err = aivmsg.UnmarshalAivSubscribeRequest(msg)
			case messages.AIV_SUBSCRIBE_RESPONSE:
				extractedStruct, err = aivmsg.UnmarshalAivSubscribeResponse(msg)
			case messages.AIV_UNSUBSCRIBE_REQUEST:
				extractedStruct, err = aivmsg.UnmarshalAivUnsubscribeRequest(msg)
			case messages.AIV_UNSUBSCRIBE_RESPONSE:
				extractedStruct, err = aivmsg.UnmarshalAivUnsubscribeResponse(msg)
			case messages.MBD_NOTIFY:
				extractedStruct, err = mbdmsg.UnmarshalMBDNotify(msg)
			case messages.MBD_SUBSCRIBE_REQUEST:
				extractedStruct, err = mbdmsg.UnmarshalMBDSubscribeRequest(msg)
			case messages.MBD_SUBSCRIBE_RESPONSE:
				extractedStruct, err = mbdmsg.UnmarshalMBDSubscribeResponse(msg)
			case messages.MBD_UNSUBSCRIBE_REQUEST:
				extractedStruct, err = mbdmsg.UnmarshalMBDUnsubscribeRequest(msg)
			case messages.MBD_UNSUBSCRIBE_RESPONSE:
				extractedStruct, err = mbdmsg.UnmarshalMBDUnsubscribeResponse(msg)
			case messages.TAS_INIT_REQUEST:
				extractedStruct, err = tasmsg.UnmarshalTasInitRequest(msg)
			case messages.TAS_INIT_RESPONSE:
				extractedStruct, err = tasmsg.UnmarshalTasInitResponse(msg)
			case messages.TAS_NOTIFY:
				extractedStruct, err = tasmsg.UnmarshalTasNotify(msg)
			case messages.TAS_SUBSCRIBE_REQUEST:
				extractedStruct, err = tasmsg.UnmarshalTasSubscribeRequest(msg)
			case messages.TAS_SUBSCRIBE_RESPONSE:
				extractedStruct, err = tasmsg.UnmarshalTasSubscribeResponse(msg)
			case messages.TAS_TA_REQUEST:
				extractedStruct, err = tasmsg.UnmarshalTasTaRequest(msg)
			case messages.TAS_TA_RESPONSE:
				extractedStruct, err = tasmsg.UnmarshalTasTaResponse(msg)
			case messages.TAS_TEARDOWN_REQUEST:
				extractedStruct, err = tasmsg.UnmarshalTasTeardownRequest(msg)
			case messages.TAS_TEARDOWN_RESPONSE:
				extractedStruct, err = tasmsg.UnmarshalTasTeardownResponse(msg)
			case messages.TAS_UNSUBSCRIBE_REQUEST:
				extractedStruct, err = tasmsg.UnmarshalTasUnsubscribeRequest(msg)
			case messages.TAS_UNSUBSCRIBE_RESPONSE:
				extractedStruct, err = tasmsg.UnmarshalTasUnsubscribeResponse(msg)
			case messages.V2X_CPM:
				extractedStruct, err = v2xmsg.UnmarshalV2XCpm(msg)
			case messages.V2X_NTM:
				extractedStruct, err = v2xmsg.UnmarshalV2XNtm(msg)

			}
			if err != nil {
				logger.Error(err.Error())
			} else {
				extractedType := fmt.Sprintf("%s", reflect.TypeOf(extractedStruct))
				logger.Info("Successfully unmarshalled JSON of type " + string(schema) + " to struct of type " + extractedType + ".")
			}
		}
	}

}
