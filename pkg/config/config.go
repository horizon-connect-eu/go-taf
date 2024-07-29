package config

import (
	"encoding/json"
	"github.com/pterm/pterm"
	"os"
)

// Configuration of the TAF, including its subcomponents.
type Configuration struct {
	Identifier    string
	Logging       Log
	Debug         Debug
	ChanBufSize   int
	Crypto        Crypto
	TAM           TAM
	Communication Communication
	TLEE          TLEE
	V2X           V2X
}

type Crypto struct {
	KeyFolder string
	Enabled   bool
}

type Debug struct {
	FixedSessionID      string
	FixedSubscriptionID string
	FixedRequestID      string
}

type Communication struct {
	Handler     string
	Kafka       Kafka
	TafEndpoint string
	AivEndpoint string
	MbdEndpoint string
}

type Kafka struct {
	Broker   string
	TafTopic string
}

type Log struct {
	LogLevel pterm.LogLevel
	LogStyle string //PLAIN, PRETTY, JSON
}

// TAMConfiguration stores the config of the [tam.tam].
type TAM struct {
	TrustModelInstanceShards int
}

type TLEE struct {
	UseInternalTLEE bool
}

type V2X struct {
	NodeTTLsec       int
	CheckIntervalSec int
}

var (
	// Default configuration of the TAF.
	// This configuration will be used if no configuration
	// file is specified explicitly by the user.
	// In case the user-specified configuration file
	// misses values, this struct defines the corresponding
	// default values.
	DefaultConfig Configuration = Configuration{
		Identifier:  "taf",
		Logging:     Log{LogLevel: pterm.LogLevelDebug, LogStyle: "PRETTY"},
		ChanBufSize: 1_000,
		TAM: TAM{
			TrustModelInstanceShards: 1,
		},
		Crypto: Crypto{
			KeyFolder: "res/cert/",
			Enabled:   false,
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
		TLEE: TLEE{
			UseInternalTLEE: true,
		},
		V2X: V2X{
			NodeTTLsec:       5,
			CheckIntervalSec: 1,
		},
	}
)

// Load a configuration from a JSON file.
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
