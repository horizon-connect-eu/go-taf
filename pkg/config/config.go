package config

import (
	"encoding/json"
	"github.com/pterm/pterm"
	"os"
)

// Configuration of the TAF, including its subcomponents.
type Configuration struct {
	ChanBufSize int
	Identifier  string

	Communication Communication
	Crypto        Crypto
	Debug         Debug
	Evidence      Evidence
	Logging       Log
	TAM           TAM
	TLEE          TLEE
	V2X           V2X
}

/*
Crypto configuration.
*/
type Crypto struct {
	KeyFolder                 string
	Enabled                   bool
	IgnoreVerificationResults bool
}

/*
Evidence-related configuration.
*/
type Evidence struct {
	AIV AIV
}

/*
AIV-related configuration.
*/
type AIV struct {
	CheckInterval int //check interval (in msec) passed to the AIV for new AIV subscriptions
}

/*
Debug configuration settings.
*/
type Debug struct {
	FixedSessionID      string // If not empty, the set session ID will be used and override all session IDs.
	FixedSubscriptionID string // If not empty, the set subscription ID will be used and override all subscription IDs.
	FixedRequestID      string // If not empty, the set request ID will be used and override all request IDs.
}

/*
Communication-related configuration.
*/
type Communication struct {
	Handler     string
	Kafka       Kafka
	TafEndpoint string
	AivEndpoint string
	MbdEndpoint string
}

/*
Kafka-related configuration.
*/
type Kafka struct {
	Broker   string //broker endpoint
	TafTopic string //Kafka topic from which the TAF consumes messages from.
}

/*
Log-related configuration.
*/
type Log struct {
	LogLevel pterm.LogLevel
	LogStyle string //PLAIN, PRETTY, JSON
}

// TAM-Configuration for settings for the Trust Assessment Manager.
type TAM struct {
	TrustModelInstanceShards int //The TAM delegates tasks to workers by partitioning all trust model instances into shards. Each shard is then backed by a single worker. This configuration parameter sets the amount of partitions/workers.
}

/*
TLEE-related configuration.
*/
type TLEE struct {
	UseInternalTLEE bool //If set to true, the TAF will use an internal debugging TLEE instead of the real TLEE for debugging purposes.
}

/*
V2X-Observer settings.
*/
type V2X struct {
	NodeTTLsec       int //This period (in sec) defines the timeout for nodes considered present by the V2XObserver.
	CheckIntervalSec int //This value (in sec) specifies in which frequency timeouts should be checked.
}

var (
	/*
	 Default configuration of the TAF.
	 This configuration will be used if no configuration
	 file is specified explicitly by the user.
	 In case the user-specified configuration file
	 misses values, this struct defines the corresponding
	 default values.
	*/
	DefaultConfig = Configuration{
		Identifier:  "taf",
		Logging:     Log{LogLevel: pterm.LogLevelDebug, LogStyle: "PRETTY"},
		ChanBufSize: 1_000,
		TAM: TAM{
			TrustModelInstanceShards: 1,
		},
		Crypto: Crypto{
			KeyFolder:                 "res/cert/",
			Enabled:                   true,
			IgnoreVerificationResults: false,
		},
		Communication: Communication{
			Handler: "kafka-based",
			Kafka: Kafka{
				Broker:   "localhost:9092",
				TafTopic: "taf",
			},
			TafEndpoint: "taf",
			AivEndpoint: "aiv",
			MbdEndpoint: "mbd",
		},
		Debug: Debug{
			FixedSessionID:      "",
			FixedSubscriptionID: "",
			FixedRequestID:      "",
		},
		Evidence: Evidence{
			AIV: AIV{
				CheckInterval: 1000,
			},
		},
		TLEE: TLEE{
			UseInternalTLEE: false,
		},
		V2X: V2X{
			NodeTTLsec:       5,
			CheckIntervalSec: 1,
		},
	}
)

/*
LoadJSON loads a configuration from a JSON file.
*/
func LoadJSON(filepath string) (Configuration, error) {
	// TODO figure out whether deep-copy is necessary here.
	config := DefaultConfig
	raw, err := os.ReadFile(filepath)
	if err != nil {
		return Configuration{}, err
	}
	err = json.Unmarshal(raw, &config)
	if err != nil {
		return Configuration{}, err
	}
	return config, nil
}
