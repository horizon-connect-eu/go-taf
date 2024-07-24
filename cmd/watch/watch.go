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
	"flag"
	"fmt"
	"github.com/IBM/sarama"
	logging "github.com/vs-uulm/go-taf/internal/logger"
	"github.com/vs-uulm/go-taf/internal/util"
	"github.com/vs-uulm/go-taf/internal/validator"
	"github.com/vs-uulm/go-taf/pkg/communication"
	"github.com/vs-uulm/go-taf/pkg/config"
	messages "github.com/vs-uulm/go-taf/pkg/message"
	"log/slog"
	"math/rand/v2"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"sync"
	"syscall"
	"time"
)

var WATCH_TOPICS = []string{"taf", "aiv", "mbd", "application.ccam", "application.migration", "application.ima", "application.smtd", "application", "tch", "v2x"}
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
	//specification of config path
	configPath := flag.String("config", "", "Path to the file with the configuration specification")
	//specification of topics to listen to
	topics := flag.Bool("topics", false, "List of space-separated topics the watcher should subscribe to.")

	flag.Parse()

	if *topics == true {
		WATCH_TOPICS = flag.Args()
	}

	tafConfig := config.DefaultConfig

	if *configPath != "" {
		var err error
		tafConfig, err = config.LoadJSON(*configPath)

		if err != nil {
			fmt.Fprintln(os.Stderr, "Config parameter is incorrect - specified file "+*configPath+" not found")
			os.Exit(1)
		}
	} else if filepath, ok := os.LookupEnv("TAF_CONFIG"); ok {
		var err error
		tafConfig, err = config.LoadJSON(filepath)
		if err != nil {
			//log.Fatalf("main: error reading config file %s: %s\n", filepath, err.Error())
			fmt.Fprintln(os.Stderr, "Environment variable is incorrect - specified file "+filepath+" not found")
			os.Exit(1)
		}
	}

	logger = logging.CreateMainLogger(tafConfig.Logging)
	logger.Info("Configuration loaded")
	logger.Debug("Running with following configuration",
		slog.String("Broker", fmt.Sprintf("%+v", tafConfig.CommunicationConfiguration.Kafka.Broker)))

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
		os.Exit(1)
	}
	defer client.Close()

	logger.Info("Starting Kafka Consumer", "Subscribed Topics", strings.Join(WATCH_TOPICS, " "))

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
		messageType := checkMessage(string(msg.Value))

		/*
			TODO: Fix AIV_RESPONSE HANDLING
		*/
		/*
			if messageType == messages.AIV_RESPONSE {
				logger.Info("Is AIV_RESPONSE")
				var MapMessage map[string]interface{}
				json.Unmarshal(msg.Value, &MapMessage)
				AivResponse, _ := json.Marshal(MapMessage["message"].(map[string]interface{})["aivEvidence"])
				trusteeReportByteStream, _ := json.Marshal(MapMessage["message"].(map[string]interface{})["trusteeReports"])
				crypto.VerifyAivResponse(AivResponse, trusteeReportByteStream, logger)
			}
		*/
		util.UNUSED(messageType)
		sess.MarkMessage(msg, "")
	}
	return nil
}

type GenericMessage struct {
	ServiceType string
	MessageType string
	Message     interface{}
}

func checkMessage(message string) messages.MessageSchema {
	var msg json.RawMessage //Placeholder for the remaining JSON later be unmarshalled using the correct type.
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

			extractedStruct, err := communication.UnmarshallMessage(schema, msg)
			if err != nil {
				logger.Error(err.Error())
			} else {
				extractedType := fmt.Sprintf("%s", reflect.TypeOf(extractedStruct))
				logger.Info("Successfully unmarshalled JSON of type " + string(schema) + " to struct of type " + extractedType + ".")
			}
		}
	}
	return schema
}
