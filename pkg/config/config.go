package config

import (
	"encoding/json"
	"github.com/pterm/pterm"
	"os"
)

// Configuration of the TAF, including its subcomponents.
type Configuration struct {
	Logging                    LogConfiguration
	ChanBufSize                int
	TAM                        TAMConfiguration
	CommunicationConfiguration CommunicationConfiguration
	TLEE                       TLEEConfig
}

type CommunicationConfiguration struct {
	Handler string
	Kafka   KafkaConfig
}

type KafkaConfig struct {
	Broker string
	Topics []string
}

type LogConfiguration struct {
	LogLevel pterm.LogLevel
	LogStyle string //PLAIN, PRETTY, JSON
}

// TAMConfiguration stores the config of the [tam.tam].
type TAMConfiguration struct {
	TrustModelInstanceShards int
}

type TLEEConfig struct {
	UseInternalTLEE bool
}

var (
	// Default configuration of the TAF.
	// This configuration will be used if no configuration
	// file is specified explicitly by the user.
	// In case the user-specified configuration file
	// misses values, this struct defines the corresponding
	// default values.
	DefaultConfig Configuration = Configuration{
		Logging:     LogConfiguration{LogLevel: pterm.LogLevelDebug, LogStyle: "PRETTY"},
		ChanBufSize: 1_000,
		TAM: TAMConfiguration{
			TrustModelInstanceShards: 1,
		},
		CommunicationConfiguration: CommunicationConfiguration{
			Handler: "kafka-based",
			Kafka: KafkaConfig{
				Broker: "localhost:9092",
				Topics: []string{"taf"},
			},
		},
		TLEE: TLEEConfig{
			UseInternalTLEE: true,
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
